package main

// Change a little bit the default kidgpin template
const DefaultUsageTemplate = `{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatCommandList"}}\
{{  range .}}\
{{    if not .Hidden}}\
{{      .Depth|Indent}}{{.Name}} {{if .Default}}*{{end}}{{template "FormatCommand" .}}\
{{      if .Commands }}<commands>
{{        .Help|Wrap 4}}\
    <commands> can be {{range .Commands}}'{{.Name}}' {{end}}. Use {{.Name}} --help for details.
{{      else}}
{{        .Help|Wrap 4}}\
{{      end}}\

{{    end}}
{{  end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .FlattenedCommands}}\
{{if not .Hidden}}\
  {{.FullCommand}}{{if .Default}}*{{end}}{{template "FormatCommand" .}}
{{.Help|Wrap 4}}
{{end}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help}}\
{{end}}\

{{end}}\

{{if .Context.SelectedCommand}}\
usage: {{.App.Name}} {{.Context.SelectedCommand}}\
{{  range  $Flagname, $FlagOpts := .Context.Flags}}\
{{    if eq $FlagOpts.Name "apps" }} --apps {{$FlagOpts.Value}}\
{{    end}}\
{{  end}}\
{{template "FormatUsage" .Context.SelectedCommand}}
{{else}}\
usage: {{.App.Name}}{{template "FormatUsage" .App}}
{{end}}\
{{if .Context.Flags}}\
Flags:
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.Args}}\
Args:
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand}}\
{{if len .Context.SelectedCommand.Commands}}\
Subcommands:
{{template "FormatCommands" .Context.SelectedCommand}}
{{end}}\
{{else if .App.Commands}}\
Commands :
{{template "FormatCommandList" .App.Commands}}
{{end}}\
`
