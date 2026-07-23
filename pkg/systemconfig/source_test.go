package systemconfig

import "testing"

func TestSourceIndexFromYAMLTracksStaticAndEnvironmentValues(t *testing.T) {
	index, err := SourceIndexFromYAML([]byte(`
bootstrap:
  mode: all
store:
  option:
    master:
      dns: ${TEST_DB_USER}:${TEST_DB_PASSWORD}@tcp($TEST_DB_HOST)/pole
`))
	if err != nil {
		t.Fatalf("SourceIndexFromYAML() error = %v", err)
	}
	if got := index.Source("bootstrap.mode"); got.Kind != SourceStaticFile {
		t.Fatalf("bootstrap.mode source = %#v", got)
	}
	dsn := index.Source("store.option.master.dns")
	if dsn.Kind != SourceEnvironment || dsn.Reference != "env:TEST_DB_USER, env:TEST_DB_PASSWORD, env:TEST_DB_HOST" {
		t.Fatalf("dsn source = %#v", dsn)
	}
	if got := index.Source("cache.diffTime"); got.Kind != SourceCompiledDefault {
		t.Fatalf("missing field source = %#v", got)
	}
}

func TestSourceIndexSubtreeMakesPathsRelative(t *testing.T) {
	index := SourceIndex{
		"bootstrap.console.webServer.listenPort": {Kind: SourceStaticFile},
		"bootstrap.mode":                         {Kind: SourceStaticFile},
	}
	console := index.Subtree("bootstrap.console")
	if got := console.Source("webServer.listenPort"); got.Kind != SourceStaticFile {
		t.Fatalf("console source = %#v", got)
	}
	if got := console.Source("bootstrap.mode"); got.Kind != SourceCompiledDefault {
		t.Fatalf("unrelated source = %#v", got)
	}
}

func TestSourceIndexSourceAggregatesListEntries(t *testing.T) {
	index, err := SourceIndexFromYAML([]byte(`
agent:
  tools:
    - list_namespaces
    - ${TEST_AGENT_TOOL}
  emptyTools: []
`))
	if err != nil {
		t.Fatalf("SourceIndexFromYAML() error = %v", err)
	}

	tools := index.Source("agent.tools")
	if tools.Kind != SourceEnvironment || tools.Reference != "env:TEST_AGENT_TOOL" {
		t.Fatalf("agent.tools source = %#v", tools)
	}
	if got := index.Source("agent.emptyTools"); got.Kind != SourceStaticFile {
		t.Fatalf("agent.emptyTools source = %#v", got)
	}
}
