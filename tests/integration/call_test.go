package integration_test

import (
	"github.com/ddelpero/gonja/v2"
	"github.com/ddelpero/gonja/v2/exec"
	"github.com/ddelpero/gonja/v2/loaders"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("call blocks", func() {
	var (
		identifier = new(string)

		environment = new(*exec.Environment)
		loader      = new(loaders.Loader)

		context = new(*exec.Context)

		returnedResult = new(string)
		returnedErr    = new(error)
		shouldRender   = func(template, result string) {
			Context(template, func() {
				BeforeEach(func() {
					*loader = loaders.MustNewMemoryLoader(map[string]string{
						*identifier: template,
					})
				})
				It("should return the expected rendered content", func() {
					By("not returning any error")
					Expect(*returnedErr).To(BeNil())
					By("returning the expected result")
					AssertPrettyDiff(result, *returnedResult)
				})
			})
		}
	)
	BeforeEach(func() {
		*identifier = "/test"
		*environment = gonja.DefaultEnvironment
		*loader = loaders.MustNewMemoryLoader(nil)
	})
	JustBeforeEach(func() {
		var t *exec.Template
		t, *returnedErr = exec.NewTemplate(*identifier, gonja.DefaultConfig, *loader, *environment)
		if *returnedErr != nil {
			return
		}
		*returnedResult, *returnedErr = t.ExecuteToString(*context)
	})

	Context("basic call block", func() {
		shouldRender(
			`{%- macro render_dialog(title, class='dialog') -%}
<div class="{{ class }}">
<h2>{{ title }}</h2>
<div class="contents">
{{- caller() }}
</div>
</div>
{%- endmacro %}
{%- call render_dialog('Hello World') %}
This is a simple dialog rendered by using a macro and a call block.
{%- endcall %}`,
			`<div class="dialog">
<h2>Hello World</h2>
<div class="contents">
This is a simple dialog rendered by using a macro and a call block.
</div>
</div>`,
		)
	})

	Context("call block with custom class", func() {
		shouldRender(
			`{%- macro render_dialog(title, class='dialog') -%}
<div class="{{ class }}">
<h2>{{ title }}</h2>
<div class="contents">
{{- caller() }}
</div>
</div>
{%- endmacro %}
{%- call render_dialog('Important', class='alert') %}
Alert content here
{%- endcall %}`,
			`<div class="alert">
<h2>Important</h2>
<div class="contents">
Alert content here
</div>
</div>`,
		)
	})

	Context("call block with caller arguments", func() {
		shouldRender(
			`{%- macro dump_users(users) -%}
<ul>
{%- for user in users %}
<li><p>{{ user }}</p>{{- caller(user) }}</li>
{%- endfor %}
</ul>
{%- endmacro %}
{%- set list_of_users = ['Alice', 'Bob', 'Charlie'] -%}
{%- call(user) dump_users(list_of_users) -%}
User: {{ user }}
{%- endcall %}`,
			`<ul>
<li><p>Alice</p>User: Alice</li>
<li><p>Bob</p>User: Bob</li>
<li><p>Charlie</p>User: Charlie</li>
</ul>`,
		)
	})

	Context("call block with multiple caller arguments", func() {
		shouldRender(
			`{%- macro table_header(cols) -%}
<table>
<tr>
{%- for col in cols %}
{{- caller(col, loop.index0) }}
{%- endfor %}
</tr>
{%- endmacro %}
{%- call(col, idx) table_header(['Name', 'Email', 'Status']) %}
<th data-col="{{ idx }}">{{ col }}</th>
{%- endcall %}
</table>`,
			`<table>
<tr>
<th data-col="0">Name</th>
<th data-col="1">Email</th>
<th data-col="2">Status</th>
</tr>
</table>`,
		)
	})

	Context("nested call blocks", func() {
		shouldRender(
			`{%- macro outer(title) -%}
<div class="outer">
<h1>{{ title }}</h1>
{{- caller() }}
</div>
{%- endmacro %}
{%- macro inner(subtitle) -%}
<div class="inner">
<h2>{{ subtitle }}</h2>
{{- caller() }}
</div>
{%- endmacro %}
{%- call outer('Main') %}
{%- call inner('Sub') %}
Content here
{%- endcall %}
{%- endcall %}`,
			`<div class="outer">
<h1>Main</h1>
<div class="inner">
<h2>Sub</h2>
Content here
</div>
</div>`,
		)
	})

	Context("call block with whitespace control", func() {
		shouldRender(
			`{%- macro render_box(title) -%}
<div class="box">
<h3>{{ title }}</h3>
{{ caller() }}
</div>
{%- endmacro -%}
{%- call render_box('Box') -%}
Box content
{%- endcall -%}`,
			`<div class="box">
<h3>Box</h3>
Box content
</div>`,
		)
	})

	Context("call block without caller in macro", func() {
		shouldRender(
			`{%- macro no_caller() -%}
<div>No caller here</div>
{%- endmacro %}
{%- call no_caller() %}
This should still work
{%- endcall %}`,
			`<div>No caller here</div>`,
		)
	})

	Context("call block with caller and loop variable", func() {
		shouldRender(
			`{%- macro list_items(items) -%}
<ul>
{%- for item in items %}
<li>{{ caller(item, loop) }}</li>
{%- endfor %}
</ul>
{%- endmacro %}
{%- call(item, loop) list_items(['a', 'b', 'c']) %}{{ loop.index }}: {{ item }}{%- endcall %}`,
			`<ul>
<li>1: a</li>
<li>2: b</li>
<li>3: c</li>
</ul>`,
		)
	})
})
