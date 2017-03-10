package grpccmd

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/protoc-gen-go/generator"
)

type grpccmd struct {
	gen *generator.Generator
}

func init() {
	generator.RegisterPlugin(new(grpccmd))
}

// Name returns the name of this plugin, "grpc".
func (g *grpccmd) Name() string {
	return "grpccmd"
}

// P forwards to g.gen.P.
func (g *grpccmd) P(args ...interface{}) { g.gen.P(args...) }

func (g *grpccmd) Init(gen *generator.Generator) {
	g.gen = gen
}
func (g *grpccmd) GenerateImports(file *generator.FileDescriptor) {
	g.P("// grpccmd imports")
	g.P(`import (
	"github.com/nathanielc/grpccmd"
	"github.com/spf13/cobra"
)`)
}

func (g *grpccmd) Generate(file *generator.FileDescriptor) {
	g.P("// Begin grpccmd ")

	g.P("var _ = grpccmd.RunE")

	for _, f := range g.gen.Request.ProtoFile {
		for _, s := range f.GetService() {
			var methodVars []string
			name := s.GetName()

			g.P("// ", name)
			serviceCmdVar := fmt.Sprintf("_%sCmd", name)
			g.P("var ", serviceCmdVar, " = &cobra.Command{")
			g.P(`Use: "`, lowerFirst(name), ` [method]",`)
			g.P(`Short: "Subcommand for the `, name, ` service.",`)
			g.P("}")
			g.P()

			for _, m := range s.GetMethod() {
				methodName := m.GetName()
				methodCmdVar := fmt.Sprintf("_%s_%sCmd", name, methodName)
				methodVars = append(methodVars, methodCmdVar)
				g.P("var ", methodCmdVar, " = &cobra.Command{")
				g.P(`Use: "`, lowerFirst(methodName), `",`)
				g.P(fmt.Sprintf(
					`Short: "Make the %s method call, input-type: %s output-type: %s",`,
					methodName,
					toTypeName(m.GetInputType()),
					toTypeName(m.GetOutputType()),
				))
				g.P(fmt.Sprintf(
					`RunE: grpccmd.RunE(
						"%s",
						"%s",
						func(c *grpc.ClientConn) interface{} {
						return New%sClient(c)
					},
				),`,
					methodName,
					toTypeName(m.GetInputType()),
					name,
				))
				g.P("}")
				g.P()

			}

			g.P("// Register commands with the root command and service command")
			g.P("func init() {")
			g.P("grpccmd.RegisterServiceCmd(", serviceCmdVar, ")")
			g.P(serviceCmdVar, ".AddCommand(")
			for _, n := range methodVars {
				g.P(n, ",")
			}
			g.P(")")
			g.P("}")
			g.P()
		}
	}

	g.P("// End grpccmd")
}

func toTypeName(t string) string {
	// Understand the correct rules here
	return strings.TrimPrefix(t, ".")
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
