//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/nucleuscloud/go-antlrv4-parser/tsql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type Input struct {
	Folder  string `json:"folder"`
	SqlFile string `json:"sql_file"`
	Driver  string `json:"driver"`
}

type Column struct {
	Name    string
	TypeStr string
}

type Table struct {
	Schema  string
	Name    string
	Columns []*Column
}

type JobMapping struct {
	Schema      string
	Table       string
	Column      string
	Transformer string
	Config      string
}

func parsePostegresStatements(sql string) ([]*Table, error) {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		return nil, err
	}

	tables := []*Table{}
	var schema string
	for _, stmt := range tree.GetStmts() {
		s := stmt.GetStmt()
		switch s.Node.(type) {
		case *pg_query.Node_CreateSchemaStmt:
			schema = s.GetCreateSchemaStmt().GetSchemaname()
		case *pg_query.Node_CreateStmt:
			table := s.GetCreateStmt().GetRelation().GetRelname()
			columns := []*Column{}
			for _, col := range s.GetCreateStmt().GetTableElts() {
				if col.GetColumnDef() != nil {
					columns = append(columns, &Column{
						Name: col.GetColumnDef().Colname,
					})
				}
			}
			tables = append(tables, &Table{
				Schema:  schema,
				Name:    table,
				Columns: columns,
			})
		}
	}
	if schema == "" {
		return nil, fmt.Errorf("unable to determine schema")
	}
	return tables, nil
}

// todo fix very brittle
func parseSQLStatements(sql string) []*Table {
	lines := strings.Split(sql, "\n")
	tableColumnsMap := make(map[string][]string)
	var currentSchema, currentTable string

	reUSE := regexp.MustCompile(`USE\s+(\w+);`)
	reCreateTable := regexp.MustCompile(`CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)\s*\.\s*(\w+)\s*\(`)
	reCreateTableNoSchema := regexp.MustCompile(`CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)\s*\(`)
	reColumn := regexp.MustCompile(`^\s*([\w]+)\s+[\w\(\)]+.*`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := reUSE.FindStringSubmatch(line); len(matches) > 1 {
			currentSchema = matches[1]
		} else if matches := reCreateTable.FindStringSubmatch(line); len(matches) > 2 {
			currentSchema = matches[1]
			currentTable = matches[2]
		} else if matches := reCreateTableNoSchema.FindStringSubmatch(line); len(matches) > 1 {
			currentTable = matches[1]
		} else if currentTable != "" {
			if matches := reColumn.FindStringSubmatch(line); len(matches) > 1 {
				columnName := matches[1]
				if slices.Contains([]string{"primary key", "constraint", "key", "unique", "primary", "alter"}, strings.ToLower(matches[1])) {
					continue
				}
				key := currentSchema + "." + currentTable
				tableColumnsMap[key] = append(tableColumnsMap[key], columnName)
			} else if strings.HasPrefix(line, "PRIMARY KEY") || strings.HasPrefix(line, "CONSTRAINT") || strings.HasPrefix(line, "UNIQUE") || strings.HasPrefix(line, "KEY") || strings.HasPrefix(line, "ENGINE") || strings.HasPrefix(line, ")") {
				// Ignore key constraints and end of table definition
				if strings.HasPrefix(line, ")") {
					currentTable = ""
				}
			}
		}
	}
	res := []*Table{}
	for table, cols := range tableColumnsMap {
		tableCols := []*Column{}
		for _, c := range cols {
			tableCols = append(tableCols, &Column{
				Name: c,
			})
		}
		split := strings.Split(table, ".")
		res = append(res, &Table{
			Schema:  split[0],
			Name:    split[1],
			Columns: tableCols,
		})
	}

	return res
}

func generateJobMapping(tables []*Table) []*mgmtv1alpha1.JobMapping {
	mappings := []*mgmtv1alpha1.JobMapping{}
	for _, t := range tables {
		for _, c := range t.Columns {
			mappings = append(mappings, &mgmtv1alpha1.JobMapping{
				Schema: t.Schema,
				Table:  t.Name,
				Column: c.Name,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
				},
			})

		}
	}
	return mappings
}

type TemplateData struct {
	SourceFile      string
	PackageName     string
	Mappings        []*mgmtv1alpha1.JobMapping
	Tables          []*Table
	GenerateTypeMap bool
}

func formatJobMappings(pkgName string, sqlFile string, mappings []*mgmtv1alpha1.JobMapping, tables []*Table, generateTypeMap bool) (string, error) {
	const tmpl = `
// Code generated by Neosync jobmapping_generator. DO NOT EDIT.
// source: {{ .SourceFile }}

package {{ .PackageName }}

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func GetDefaultSyncJobMappings()[]*mgmtv1alpha1.JobMapping {
  return []*mgmtv1alpha1.JobMapping{
		{{- range .Mappings }}
		{
			Schema: "{{ .Schema }}",
			Table:  "{{ .Table }}",
			Column: "{{ .Column }}",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
			},
		},
		{{- end }}
	} 
}
{{ if .GenerateTypeMap }}


func GetTableColumnTypeMap() map[string]map[string]string {
	return map[string]map[string]string{
		{{- range .Tables }}
		"{{ .Schema }}.{{ .Name }}": {
		{{- range .Columns }}
			"{{ .Name }}": "{{ .TypeStr }}",
		{{- end }}
		},
		{{- end }}
	}
}
{{- end }}
`
	data := TemplateData{
		SourceFile:      sqlFile,
		PackageName:     pkgName,
		Mappings:        mappings,
		Tables:          tables,
		GenerateTypeMap: generateTypeMap,
	}
	t := template.Must(template.New("jobmappings").Parse(tmpl))
	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func main() {
	args := os.Args
	if len(args) < 3 {
		panic("must provide necessary args")
	}

	configFile := args[1]
	gopackage := args[2]

	packageSplit := strings.Split(gopackage, "_")
	goPkg := packageSplit[len(packageSplit)-1]

	jsonFile, err := os.Open(configFile)
	if err != nil {
		fmt.Println("failed to open file: %s", err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("failed to read file: %s", err)
		return
	}

	var inputs []Input
	if err := json.Unmarshal(byteValue, &inputs); err != nil {
		fmt.Println("failed to unmarshal JSON: %s", err)
		return
	}
	for _, input := range inputs {
		folderSplit := strings.Split(input.Folder, "/")
		var goPkgName string
		if len(folderSplit) == 1 {
			goPkgName = strings.ReplaceAll(fmt.Sprintf("%s_%s", goPkg, input.Folder), "-", "")
		} else if len(folderSplit) > 1 {
			lastTwo := folderSplit[len(folderSplit)-2:]
			goPkgName = strings.ReplaceAll(strings.Join(lastTwo, "_"), "-", "")
		}
		sqlFile, err := os.Open(fmt.Sprintf("%s/%s", input.Folder, input.SqlFile))
		if err != nil {
			fmt.Println("failed to open file: %s", err)
		}

		byteValue, err := io.ReadAll(sqlFile)
		if err != nil {
			fmt.Println("failed to read file: %s", err)
		}

		sqlContent := string(byteValue)
		sqlFile.Close()

		var tables []*Table
		if input.Driver == "postgres" {
			t, err := parsePostegresStatements(sqlContent)
			if err != nil {
				fmt.Println("Error parsing postgres SQL schema:", err)
				return
			}
			tables = t
		} else if input.Driver == "mysql" {
			t := parseSQLStatements(sqlContent)
			tables = t
		} else if input.Driver == "sqlserver" {
			t := parseTsql(sqlContent)
			tables = t
		}

		jobMapping := generateJobMapping(tables)

		formattedJobMappings, err := formatJobMappings(goPkgName, input.SqlFile, jobMapping, tables, input.Driver == "sqlserver")
		if err != nil {
			fmt.Println("Error formatting job mappings:", err)
			return
		}

		output := fmt.Sprintf("%s/job_mappings.go", input.Folder)
		outputFile, err := os.Create(output)
		if err != nil {
			fmt.Println("Error creating jobmapping.go file:", err)
			return
		}

		_, err = outputFile.WriteString(formattedJobMappings)
		if err != nil {
			fmt.Println("Error writing to jobmapping.go file:", err)
			return
		}
		outputFile.Close()
	}

	return
}

type tsqlListener struct {
	*parser.BaseTSqlParserListener
	inCreate      bool
	currentSchema string
	currentTable  string
	currentCols   []*Column
	mappings      []*Table
}

func (l *tsqlListener) PushTable() {
	l.mappings = append(l.mappings, &Table{
		Schema:  l.currentSchema,
		Name:    l.currentTable,
		Columns: l.currentCols,
	})
	l.currentSchema = ""
	l.currentTable = ""
	l.currentCols = []*Column{}
	l.inCreate = false
}

func (l *tsqlListener) PushColumn(name, typeStr string) {
	l.currentCols = append(l.currentCols, &Column{
		Name:    name,
		TypeStr: typeStr,
	})
}

func (l *tsqlListener) SetTable(schemaTable string) {
	split := strings.Split(schemaTable, ".")
	if len(split) == 1 {
		l.currentSchema = "dbo"
		l.currentTable = split[0]
	} else if len(split) > 1 {
		l.currentSchema = split[0]
		l.currentTable = split[1]
	}
}

// EnterCreate_table is called when production create_table is entered.
func (l *tsqlListener) EnterCreate_table(ctx *parser.Create_tableContext) {
	l.inCreate = true
	table := ctx.Table_name().GetText()
	l.SetTable(table)
}

// ExitCreate_table is called when production create_table is exited.
func (l *tsqlListener) ExitCreate_table(ctx *parser.Create_tableContext) {
	l.PushTable()
}
func (l *tsqlListener) EnterColumn_definition(ctx *parser.Column_definitionContext) {
	l.PushColumn(ctx.Id_().GetText(), ctx.Data_type().GetText())
}

func parseTsql(sql string) []*Table {
	inputStream := antlr.NewInputStream(sql)

	// create the lexer
	lexer := parser.NewTSqlLexer(inputStream)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// create the parser
	p := parser.NewTSqlParser(tokens)

	listener := &tsqlListener{}
	tree := p.Tsql_file()
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	return listener.mappings
}
