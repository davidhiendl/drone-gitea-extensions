package plugin_converter

import (
	"dhswt.de/drone-gitea-extensions/shared"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlugin(t *testing.T) {

}

func testPluginYamlInclude(mainYaml string, includeTestYaml string, expectedYaml string, t *testing.T) {

	// generate a test server so we can capture and inspect the request
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/test.yaml" {
			res.WriteHeader(200)
			res.Write([]byte(includeTestYaml))
		} else {
			res.WriteHeader(404)
		}
	}))
	defer func() { testServer.Close() }()

	plugin := New(nil, &shared.AppConfig{
		GiteaURL:              "https://gitea.example.com/",
		DroneConfigIncludeMax: 10,
	}, nil)

	expectedYaml = strings.ReplaceAll(expectedYaml, "__TEST_SERVER_URL__", testServer.URL)
	mainYaml = strings.ReplaceAll(mainYaml, "__TEST_SERVER_URL__", testServer.URL)

	mergedYaml, err := plugin.regexReplaceIncludeDirectives(mainYaml, http.DefaultClient)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// trim whitespace
	//compareCutset := "\n\r "
	//mergedYaml = strings.Trim(mergedYaml, compareCutset)
	//expectedYaml = strings.Trim(expectedYaml, compareCutset)

	// make whitespace visible
	//mergedYaml = strings.ReplaceAll(mergedYaml, " ", "%")
	//expectedYaml = strings.ReplaceAll(expectedYaml, " ", "%")

	fmt.Printf("=== MERGED YAML\n%s\n=== END OF MERGED YAML\n", mergedYaml)
	fmt.Printf("=== EXPECTED YAML\n%s\n=== END OF EXPECTED YAML\n", expectedYaml)

	if expectedYaml != mergedYaml {
		t.Fatalf("merged yaml does not equal expected yaml")
	}
}
func TestPluginYamlIncludeBasic(t *testing.T) {

	includeTestYaml := `
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
`

	mainYaml := `
# test comment
testKey:SomeValue
_include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	expectedYaml := `
# test comment
testKey:SomeValue
# DIRECTIVE_START _include: __TEST_SERVER_URL__/test.yaml
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
# DIRECTIVE_END _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	testPluginYamlInclude(mainYaml, includeTestYaml, expectedYaml, t)

}

func TestPluginYamlIncludeMultiple(t *testing.T) {

	includeTestYaml := `
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
`

	mainYaml := `
# test comment
testKey:SomeValue
_include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
_include: __TEST_SERVER_URL__/test.yaml
testKey5:SomeValue
`

	expectedYaml := `
# test comment
testKey:SomeValue
# DIRECTIVE_START _include: __TEST_SERVER_URL__/test.yaml
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
# DIRECTIVE_END _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
# DIRECTIVE_START _include: __TEST_SERVER_URL__/test.yaml
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
# DIRECTIVE_END _include: __TEST_SERVER_URL__/test.yaml
testKey5:SomeValue
`
	testPluginYamlInclude(mainYaml, includeTestYaml, expectedYaml, t)

}

func TestPluginYamlIncludeNotNested(t *testing.T) {

	includeTestYaml := `
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
`

	mainYaml := `
# test comment
testKey:SomeValue
parentKey:
    _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	expectedYaml := `
# test comment
testKey:SomeValue
parentKey:
    _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	testPluginYamlInclude(mainYaml, includeTestYaml, expectedYaml, t)

}

func TestPluginYamlIncludeNotIndented(t *testing.T) {

	includeTestYaml := `
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
`

	mainYaml := `
# test comment
testKey:SomeValue
 _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	expectedYaml := `
# test comment
testKey:SomeValue
 _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	testPluginYamlInclude(mainYaml, includeTestYaml, expectedYaml, t)

}

func TestPluginYamlIncludeNotCommented(t *testing.T) {

	includeTestYaml := `
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
`

	mainYaml := `
# test comment
testKey:SomeValue
// _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	expectedYaml := `
# test comment
testKey:SomeValue
// _include: __TEST_SERVER_URL__/test.yaml
testKey2:SomeValue
`
	testPluginYamlInclude(mainYaml, includeTestYaml, expectedYaml, t)

}

func TestPluginYamlIncludeNoIncludes(t *testing.T) {

	includeTestYaml := `
# test comment 1
testKey3:SomeValue
testKey4:SomeValue
`

	mainYaml := `
# test comment
testKey:SomeValue
testKey2:SomeValue
`
	expectedYaml := `
# test comment
testKey:SomeValue
testKey2:SomeValue
`
	testPluginYamlInclude(mainYaml, includeTestYaml, expectedYaml, t)

}
