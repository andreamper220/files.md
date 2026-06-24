package life

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListAllAreas_IncludesNestedSections(t *testing.T) {
	r := require.New(t)
	fsys := newTestFS(t)
	r.NoError(Init(fsys))

	projectPath, err := CreateProject(fsys, SpherePath("Работа"), "Root")
	r.NoError(err)
	sectionPath, err := CreateSection(fsys, projectPath, "Inner")
	r.NoError(err)

	areas, err := ListAllAreas(fsys, SpherePath("Работа"))
	r.NoError(err)
	r.Contains(areas, projectPath)
	r.Contains(areas, sectionPath)
	r.Equal(2, AreaDepth(sectionPath))
}

func TestAreaFullTitle_NestedPath(t *testing.T) {
	r := require.New(t)
	title := AreaFullTitle("spheres/Работа/Root/Inner")
	r.Equal("Root / Inner", title)
}
