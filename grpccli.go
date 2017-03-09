package grpccli

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/protoc-gen-go/generator"
)

type grpccli struct {
	gen *generator.Generator
}

func init() {
	generator.RegisterPlugin(new(grpccli))
}

// Name returns the name of this plugin, "grpc".
func (g *grpccli) Name() string {
	return "grpccli"
}

// P forwards to g.gen.P.
func (g *grpccli) P(args ...interface{}) { g.gen.P(args...) }

func (g *grpccli) Init(gen *generator.Generator) {
	g.gen = gen
}
func (g *grpccli) GenerateImports(file *generator.FileDescriptor) {
	g.P("// Begin grpccli imports")
	g.P(`import (
	"github.com/spf13/cobra"
	"github.com/nathanielc/grpccli"
)`)
	g.P("// End grpccli imports")
}

func (g *grpccli) Generate(file *generator.FileDescriptor) {
	g.P("// Begin grpccli ")

	g.P("var _ = grpccli.RunE")
	g.P("var _grpcAddr = new(string)")

	g.P(`
	var RootCmd = &cobra.Command{
		Use: "grpccli [command]",
		Short: "A CLI generator for gRPC services using Cobra",
	}
`)

	g.P("func init() {")
	g.P(`RootCmd.PersistentFlags().StringVar(_grpcAddr, "addr", "", "gRPC server address")`)
	g.P("}")
	g.P()

	for _, f := range g.gen.Request.ProtoFile {
		for _, s := range f.GetService() {
			var methodCmds []string
			name := s.GetName()

			g.P("// ", name)
			serviceCmdName := fmt.Sprintf("_%sCmd", name)
			g.P("var ", serviceCmdName, " = &cobra.Command{")
			g.P(`Use: "`, lowerFirst(name), ` [method]",`)
			g.P(`Short: "A CLI for the proto service `, name, `",`)
			g.P("}")
			g.P()

			for _, m := range s.GetMethod() {
				methodName := m.GetName()
				methodCmdName := fmt.Sprintf("_%s_%sCmd", name, methodName)
				methodCmds = append(methodCmds, methodCmdName)
				g.P("var ", methodCmdName, " = &cobra.Command{")
				g.P(`Use: "`, lowerFirst(methodName), `",`)
				g.P(`Short: "Make the `, methodName, ` call on the service`, name, `",`)
				g.P(fmt.Sprintf(
					`RunE: grpccli.RunE(
						_grpcAddr, 
						"%s",
						"%s",
						"%s",
						func(c *grpc.ClientConn) interface{} {
						return New%sClient(c)
					},
				),`,
					methodName,
					toTypeName(m.GetInputType()),
					toTypeName(m.GetOutputType()),
					name,
				))
				g.P("}")
				g.P()

			}

			g.P("// Register commands with the root command and service command")
			g.P("func init() {")
			g.P("RootCmd.AddCommand(", serviceCmdName, ")")
			g.P(serviceCmdName, ".AddCommand(")
			for _, n := range methodCmds {
				g.P(n, ",")
			}
			g.P(")")
			g.P("}")
			g.P()
		}
	}

	g.P("// End grpccli")
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
