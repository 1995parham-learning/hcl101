package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

var schemas = map[string]*hcl.BodySchema{
	"": {
		Attributes: []hcl.AttributeSchema{
			{Name: "version"},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "local"},
			{Type: "person", LabelNames: []string{"name"}},
		},
	},
	"person": {
		Attributes: []hcl.AttributeSchema{
			{Name: "date", Required: true},
			{Name: "birthday", Required: true},
		},
	},
}

func print(v map[string]cty.Value) {
	for name, value := range v {
		if value.Type().IsObjectType() {
			fmt.Printf("%s:\n", name)
			print(value.AsValueMap())
		} else {
			fmt.Printf("%s = %s\n", name, value.GoString())
		}
	}
}

func parse(ctx *hcl.EvalContext, body hcl.Body, t string, strict bool) {
	if schemas[t] == nil {
		attributes, _ := body.JustAttributes()
		for _, attribute := range attributes {
			name := attribute.Name
			val, diag := attribute.Expr.Value(ctx)
			if diag != nil && strict {
				log.Fatalf("fail to read the attribute %s (%s)", name, diag.Error())
			}
			if diag == nil {
				if t != "" {
					var vars map[string]cty.Value
					if ctx.Variables[t].IsNull() {
						vars = make(map[string]cty.Value)
					} else {
						vars = ctx.Variables[t].AsValueMap()
					}
					vars[name] = val
					ctx.Variables[t] = cty.ObjectVal(vars)
				} else {
					ctx.Variables[name] = val
				}
			}
		}
	} else {
		content, diag := body.Content(schemas[t])
		if diag != nil {
			log.Fatalf("fail to read the hcl content (%s)", diag.Error())
		}

		for _, attribute := range content.Attributes {
			name := attribute.Name
			val, diag := attribute.Expr.Value(ctx)
			if diag != nil && strict {
				log.Fatalf("fail to read the attribute %s (%s)", name, diag.Error())
			}
			if diag == nil {
				if t != "" {
					var vars map[string]cty.Value
					if ctx.Variables[t].IsNull() {
						vars = make(map[string]cty.Value)
					} else {
						vars = ctx.Variables[t].AsValueMap()
					}
					vars[name] = val
					ctx.Variables[t] = cty.ObjectVal(vars)
				} else {
					ctx.Variables[name] = val
				}
			}
		}

		for _, block := range content.Blocks {
			var name string
			if block.Type == "person" {
				name = strings.ReplaceAll(strings.ToLower(block.Labels[0]), " ", "_")
			} else {
				name = block.Type
			}

			parse(ctx, block.Body, name, strict)
		}
	}
}

func main() {
	parser := hclparse.NewParser()
	f, diag := parser.ParseHCLFile("family.hcl")
	if diag != nil {
		log.Fatalf("fail to read the hcl file %s", diag.Error())
	}

	ctx := new(hcl.EvalContext)
	ctx.Variables = make(map[string]cty.Value)

	parse(ctx, f.Body, "", false)
	parse(ctx, f.Body, "", true)

	print(ctx.Variables)
}
