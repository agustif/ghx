package coverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// GraphQLIntrospectionQuery is the schema query used by `ghx mine github`.
const GraphQLIntrospectionQuery = `
query GhxMineGraphQLSchema {
  __schema {
    queryType {
      name
      fields(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
        args {
          name
          description
          defaultValue
          type { ...GhxTypeRef }
        }
        type { ...GhxTypeRef }
      }
    }
    mutationType {
      name
      fields(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
        args {
          name
          description
          defaultValue
          type { ...GhxTypeRef }
        }
        type { ...GhxTypeRef }
      }
    }
    subscriptionType { name }
    types {
      kind
      name
    }
  }
}

fragment GhxTypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType { kind name }
      }
    }
  }
}`

// GraphQLFilter selects GraphQL fields in a generated report.
type GraphQLFilter struct {
	State string
	Query string
}

// GraphQLReport describes generated GraphQL coverage.
type GraphQLReport struct {
	Source              string         `json:"source"`
	Host                string         `json:"host"`
	SchemaHash          string         `json:"schemaHash,omitempty"`
	QueryType           string         `json:"queryType,omitempty"`
	MutationType        string         `json:"mutationType,omitempty"`
	SubscriptionType    string         `json:"subscriptionType,omitempty"`
	TotalTypes          int            `json:"totalTypes"`
	ObjectTypes         int            `json:"objectTypes"`
	InputTypes          int            `json:"inputTypes"`
	EnumTypes           int            `json:"enumTypes"`
	InterfaceTypes      int            `json:"interfaceTypes"`
	UnionTypes          int            `json:"unionTypes"`
	TotalFields         int            `json:"totalFields"`
	QueryFields         int            `json:"queryFields"`
	MutationFields      int            `json:"mutationFields"`
	DeprecatedFields    int            `json:"deprecatedFields"`
	TrackedFields       int            `json:"trackedFields"`
	MatchingFields      int            `json:"matchingFields"`
	ExplicitCoveragePct float64        `json:"explicitCoveragePct"`
	RemainingGapPct     float64        `json:"remainingGapPct"`
	StateCounts         []BucketCount  `json:"stateCounts"`
	Fields              []GraphQLField `json:"fields"`
}

// GraphQLField is one generated GraphQL query or mutation coverage row.
type GraphQLField struct {
	Coordinate      string            `json:"coordinate"`
	ParentType      string            `json:"parentType"`
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Description     string            `json:"description,omitempty"`
	ReturnType      string            `json:"returnType,omitempty"`
	Args            []GraphQLArgument `json:"args,omitempty"`
	Deprecated      bool              `json:"deprecated"`
	DeprecationNote string            `json:"deprecationNote,omitempty"`
	CoverageState   string            `json:"coverageState"`
	ProposedCommand string            `json:"proposedCommand,omitempty"`
	RawCommand      string            `json:"rawCommand,omitempty"`
}

// GraphQLArgument is a GraphQL field argument.
type GraphQLArgument struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
}

// GraphQLSchema is the subset of GraphQL introspection used for coverage.
type GraphQLSchema struct {
	QueryType        *GraphQLRootType   `json:"queryType"`
	MutationType     *GraphQLRootType   `json:"mutationType"`
	SubscriptionType *GraphQLRootType   `json:"subscriptionType"`
	Types            []GraphQLType      `json:"types"`
	Directives       []GraphQLDirective `json:"directives"`
}

// GraphQLRootType names a GraphQL root type.
type GraphQLRootType struct {
	Name   string          `json:"name"`
	Fields []GraphQLMember `json:"fields"`
}

// GraphQLType is a GraphQL introspection type.
type GraphQLType struct {
	Kind          string              `json:"kind"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Fields        []GraphQLMember     `json:"fields"`
	InputFields   []GraphQLInputValue `json:"inputFields"`
	Interfaces    []GraphQLTypeRef    `json:"interfaces"`
	PossibleTypes []GraphQLTypeRef    `json:"possibleTypes"`
	EnumValues    []GraphQLEnumValue  `json:"enumValues"`
}

// GraphQLMember is a GraphQL field member.
type GraphQLMember struct {
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	Args              []GraphQLInputValue `json:"args"`
	Type              GraphQLTypeRef      `json:"type"`
	IsDeprecated      bool                `json:"isDeprecated"`
	DeprecationReason string              `json:"deprecationReason"`
}

// GraphQLInputValue is a GraphQL argument or input field.
type GraphQLInputValue struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	DefaultValue string         `json:"defaultValue"`
	Type         GraphQLTypeRef `json:"type"`
}

// GraphQLEnumValue is a GraphQL enum value.
type GraphQLEnumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

// GraphQLDirective is a GraphQL directive.
type GraphQLDirective struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Locations   []string            `json:"locations"`
	Args        []GraphQLInputValue `json:"args"`
}

// GraphQLTypeRef is a nested GraphQL type reference.
type GraphQLTypeRef struct {
	Kind   string          `json:"kind"`
	Name   string          `json:"name"`
	OfType *GraphQLTypeRef `json:"ofType"`
}

// BuildGraphQLReport converts GraphQL introspection into a coverage report.
func BuildGraphQLReport(host string, schema GraphQLSchema, filter GraphQLFilter) (*GraphQLReport, error) {
	if !ValidCoverageState(filter.State) {
		return nil, fmt.Errorf("unknown coverage state %q", filter.State)
	}

	typeByName := map[string]GraphQLType{}
	totalFields := 0
	deprecatedFields := 0
	objectTypes := 0
	inputTypes := 0
	enumTypes := 0
	interfaceTypes := 0
	unionTypes := 0
	totalTypes := 0
	for _, t := range schema.Types {
		if strings.HasPrefix(t.Name, "__") {
			continue
		}
		totalTypes++
		typeByName[t.Name] = t
		totalFields += len(t.Fields)
		for _, field := range t.Fields {
			if field.IsDeprecated {
				deprecatedFields++
			}
		}
		switch t.Kind {
		case "OBJECT":
			objectTypes++
		case "INPUT_OBJECT":
			inputTypes++
		case "ENUM":
			enumTypes++
		case "INTERFACE":
			interfaceTypes++
		case "UNION":
			unionTypes++
		}
	}

	queryName := rootTypeName(schema.QueryType)
	mutationName := rootTypeName(schema.MutationType)
	subscriptionName := rootTypeName(schema.SubscriptionType)
	queryRoot := rootGraphQLType(schema.QueryType, typeByName)
	mutationRoot := rootGraphQLType(schema.MutationType, typeByName)

	if queryName != "" && len(typeByName[queryName].Fields) == 0 {
		totalFields += len(queryRoot.Fields)
		deprecatedFields += countDeprecatedGraphQLFields(queryRoot.Fields)
	}
	if mutationName != "" && mutationName != queryName && len(typeByName[mutationName].Fields) == 0 {
		totalFields += len(mutationRoot.Fields)
		deprecatedFields += countDeprecatedGraphQLFields(mutationRoot.Fields)
	}

	allFields := append(
		graphQLRootFields(queryRoot, "query"),
		graphQLRootFields(mutationRoot, "mutation")...,
	)
	sort.Slice(allFields, func(i, j int) bool {
		if allFields[i].Kind == allFields[j].Kind {
			return allFields[i].Coordinate < allFields[j].Coordinate
		}
		return allFields[i].Kind < allFields[j].Kind
	})

	var matches []GraphQLField
	for _, field := range allFields {
		if graphQLFieldMatchesFilter(field, filter) {
			matches = append(matches, field)
		}
	}

	trackedFields := 0
	for _, field := range allFields {
		if field.CoverageState != StateMissing {
			trackedFields++
		}
	}

	return &GraphQLReport{
		Source:              "graphql",
		Host:                host,
		SchemaHash:          schemaHash(schema),
		QueryType:           queryName,
		MutationType:        mutationName,
		SubscriptionType:    subscriptionName,
		TotalTypes:          totalTypes,
		ObjectTypes:         objectTypes,
		InputTypes:          inputTypes,
		EnumTypes:           enumTypes,
		InterfaceTypes:      interfaceTypes,
		UnionTypes:          unionTypes,
		TotalFields:         totalFields,
		QueryFields:         len(queryRoot.Fields),
		MutationFields:      len(mutationRoot.Fields),
		DeprecatedFields:    deprecatedFields,
		TrackedFields:       trackedFields,
		MatchingFields:      len(matches),
		ExplicitCoveragePct: percent(trackedFields, len(allFields)),
		RemainingGapPct:     percent(len(allFields)-trackedFields, len(allFields)),
		StateCounts:         bucketCounts(matches, func(field GraphQLField) string { return field.CoverageState }),
		Fields:              matches,
	}, nil
}

func countDeprecatedGraphQLFields(fields []GraphQLMember) int {
	count := 0
	for _, field := range fields {
		if field.IsDeprecated {
			count++
		}
	}
	return count
}

func rootTypeName(root *GraphQLRootType) string {
	if root == nil {
		return ""
	}
	return root.Name
}

func rootGraphQLType(root *GraphQLRootType, types map[string]GraphQLType) GraphQLType {
	if root == nil {
		return GraphQLType{}
	}
	if len(root.Fields) > 0 {
		return GraphQLType{Name: root.Name, Kind: "OBJECT", Fields: root.Fields}
	}
	return types[root.Name]
}

func graphQLRootFields(t GraphQLType, kind string) []GraphQLField {
	if t.Name == "" {
		return nil
	}
	fields := make([]GraphQLField, 0, len(t.Fields))
	for _, field := range t.Fields {
		coordinate := t.Name + "." + field.Name
		fields = append(fields, GraphQLField{
			Coordinate:      coordinate,
			ParentType:      t.Name,
			Name:            field.Name,
			Kind:            kind,
			Description:     field.Description,
			ReturnType:      formatGraphQLType(field.Type),
			Args:            graphQLArguments(field.Args),
			Deprecated:      field.IsDeprecated,
			DeprecationNote: field.DeprecationReason,
			CoverageState:   StateMissing,
			ProposedCommand: proposedGraphQLCommand(kind, field.Name),
			RawCommand:      "ghx api graphql",
		})
	}
	return fields
}

func graphQLArguments(inputs []GraphQLInputValue) []GraphQLArgument {
	args := make([]GraphQLArgument, 0, len(inputs))
	for _, input := range inputs {
		args = append(args, GraphQLArgument{
			Name:         input.Name,
			Type:         formatGraphQLType(input.Type),
			Required:     input.Type.Kind == "NON_NULL",
			DefaultValue: input.DefaultValue,
			Description:  input.Description,
		})
	}
	sort.Slice(args, func(i, j int) bool {
		return args[i].Name < args[j].Name
	})
	return args
}

func formatGraphQLType(ref GraphQLTypeRef) string {
	switch ref.Kind {
	case "NON_NULL":
		return formatGraphQLTypeRef(ref.OfType) + "!"
	case "LIST":
		return "[" + formatGraphQLTypeRef(ref.OfType) + "]"
	default:
		if ref.Name != "" {
			return ref.Name
		}
		return ref.Kind
	}
}

func formatGraphQLTypeRef(ref *GraphQLTypeRef) string {
	if ref == nil {
		return ""
	}
	return formatGraphQLType(*ref)
}

func schemaHash(schema GraphQLSchema) string {
	b, err := json.Marshal(schema)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func proposedGraphQLCommand(kind, fieldName string) string {
	name := strings.ToLower(fieldName)
	switch {
	case strings.Contains(name, "pullrequest") || strings.Contains(name, "mergequeue"):
		return "ghx pr"
	case strings.Contains(name, "repository") || strings.Contains(name, "ruleset"):
		return "ghx repo"
	case strings.Contains(name, "organization") || strings.Contains(name, "enterprise") || strings.Contains(name, "team"):
		return "ghx org"
	case strings.Contains(name, "project"):
		return "ghx board"
	case strings.Contains(name, "security") || strings.Contains(name, "vulnerabil") || strings.Contains(name, "advisory"):
		return "ghx sec"
	case strings.Contains(name, "check") || strings.Contains(name, "workflow") || strings.Contains(name, "deployment"):
		return "ghx checks"
	default:
		return fmt.Sprintf("ghx api graphql # %s %s", kind, fieldName)
	}
}

func graphQLFieldMatchesFilter(field GraphQLField, filter GraphQLFilter) bool {
	if filter.State != "" && !strings.EqualFold(filter.State, field.CoverageState) {
		return false
	}
	query := strings.ToLower(filter.Query)
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		field.Coordinate,
		field.ParentType,
		field.Name,
		field.Kind,
		field.Description,
		field.ReturnType,
		field.CoverageState,
		field.ProposedCommand,
	}, " "))
	return strings.Contains(haystack, query)
}
