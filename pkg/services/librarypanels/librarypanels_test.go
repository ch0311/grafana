package librarypanels

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/macaron.v1"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
)

func TestCreateLibraryPanel(t *testing.T) {
	testScenario(t, "When an admin tries to create a library panel that already exists, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			response = sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 400, response.Status())
		})
}

func TestDeleteLibraryPanel(t *testing.T) {
	testScenario(t, "When an admin tries to delete a library panel that does not exist, it should fail",
		func(t *testing.T, sc scenarioContext) {
			response := sc.service.deleteHandler(sc.reqContext)
			require.Equal(t, 404, response.Status())
		})

	testScenario(t, "When an admin tries to delete a library panel that exists, it should succeed",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			response = sc.service.deleteHandler(sc.reqContext)
			require.Equal(t, 200, response.Status())
		})

	testScenario(t, "When an admin tries to delete a library panel in another org, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			sc.reqContext.SignedInUser.OrgId = 2
			sc.reqContext.SignedInUser.OrgRole = models.ROLE_ADMIN
			response = sc.service.deleteHandler(sc.reqContext)
			require.Equal(t, 404, response.Status())
		})
}

func TestGetLibraryPanel(t *testing.T) {
	testScenario(t, "When an admin tries to get a library panel that does not exist, it should fail",
		func(t *testing.T, sc scenarioContext) {
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": "unknown"})
			response := sc.service.getHandler(sc.reqContext)
			require.Equal(t, 404, response.Status())
		})

	testScenario(t, "When an admin tries to get a library panel that exists, it should succeed and return correct result",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			response = sc.service.getHandler(sc.reqContext)
			require.Equal(t, 200, response.Status())
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, int64(1), result.Result.FolderID)
			require.Equal(t, "Text - Library Panel", result.Result.Name)
		})

	testScenario(t, "When an admin tries to get a library panel that exists in an other org, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			sc.reqContext.SignedInUser.OrgId = 2
			sc.reqContext.SignedInUser.OrgRole = models.ROLE_ADMIN
			response = sc.service.getHandler(sc.reqContext)
			require.Equal(t, 404, response.Status())
		})
}

func TestGetAllLibraryPanels(t *testing.T) {
	testScenario(t, "When an admin tries to get all library panels and none exists, it should return none",
		func(t *testing.T, sc scenarioContext) {
			response := sc.service.getAllHandler(sc.reqContext)
			require.Equal(t, 200, response.Status())

			var result libraryPanelsResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.NotNil(t, result.Result)
			require.Equal(t, 0, len(result.Result))
		})

	testScenario(t, "When an admin tries to get all library panels and two exist, it should work",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			command = getCreateCommand(1, "Text - Library Panel2")
			response = sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			response = sc.service.getAllHandler(sc.reqContext)
			require.Equal(t, 200, response.Status())

			var result libraryPanelsResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, 2, len(result.Result))
			require.Equal(t, int64(1), result.Result[0].FolderID)
			require.Equal(t, "Text - Library Panel", result.Result[0].Name)
			require.Equal(t, int64(1), result.Result[1].FolderID)
			require.Equal(t, "Text - Library Panel2", result.Result[1].Name)
		})

	testScenario(t, "When an admin tries to get all library panels in a different org, none should be returned",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")

			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			response = sc.service.getAllHandler(sc.reqContext)
			require.Equal(t, 200, response.Status())

			var result libraryPanelsResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, 1, len(result.Result))
			require.Equal(t, int64(1), result.Result[0].FolderID)
			require.Equal(t, "Text - Library Panel", result.Result[0].Name)

			sc.reqContext.SignedInUser.OrgId = 2
			sc.reqContext.SignedInUser.OrgRole = models.ROLE_ADMIN
			response = sc.service.getAllHandler(sc.reqContext)
			require.Equal(t, 200, response.Status())

			result = libraryPanelsResult{}
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.NotNil(t, result.Result)
			require.Equal(t, 0, len(result.Result))
		})
}

func TestUpdateLibraryPanel(t *testing.T) {
	testScenario(t, "When an admin tries to update a library panel that does not exist, it should fail",
		func(t *testing.T, sc scenarioContext) {
			cmd := updateLibraryPanelCommand{}
			response := sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 404, response.Status())
		})

	testScenario(t, "When an admin tries to update a library panel that exists, it should succeed",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				FolderID: 2,
				Name:     "Panel - New name",
				Model: []byte(`
								{
								  "datasource": "${DS_GDEV-TESTDATA}",
								  "id": 1,
								  "name": "Model - New name",
								  "type": "text"
								}
							`),
			}
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 200, response.Status())
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, int64(2), result.Result.FolderID)
			require.Equal(t, "Panel - New name", result.Result.Name)
			require.Equal(t, "Model - New name", result.Result.Model["name"])
		})

	testScenario(t, "When an admin tries to update a library panel with folder only, it should change folder successfully and return correct result",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var existing libraryPanelResult
			err := json.Unmarshal(response.Body(), &existing)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				FolderID: 100,
			}
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": existing.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 200, response.Status())
			var result libraryPanelResult
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, existing.Result.ID, result.Result.ID)
			require.Equal(t, existing.Result.OrgID, result.Result.OrgID)
			require.Equal(t, int64(100), result.Result.FolderID)
			require.Equal(t, existing.Result.Name, result.Result.Name)
			require.Equal(t, existing.Result.Model["name"], result.Result.Model["name"])
			require.Equal(t, existing.Result.Created.UTC().Format(time.RFC3339), result.Result.Created.UTC().Format(time.RFC3339))
			require.Equal(t, existing.Result.CreatedBy, result.Result.CreatedBy)
			require.True(t, existing.Result.Updated.UTC().Before(result.Result.Updated.UTC()))
			require.Equal(t, existing.Result.UpdatedBy, result.Result.UpdatedBy)
		})

	testScenario(t, "When an admin tries to update a library panel with name only, it should change name successfully and return correct result",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var existing libraryPanelResult
			err := json.Unmarshal(response.Body(), &existing)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				Name: "New Name",
			}
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": existing.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 200, response.Status())
			var result libraryPanelResult
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, existing.Result.ID, result.Result.ID)
			require.Equal(t, existing.Result.OrgID, result.Result.OrgID)
			require.Equal(t, int64(1), result.Result.FolderID)
			require.Equal(t, "New Name", result.Result.Name)
			require.Equal(t, existing.Result.Model["name"], result.Result.Model["name"])
			require.Equal(t, existing.Result.Created.UTC().Format(time.RFC3339), result.Result.Created.UTC().Format(time.RFC3339))
			require.Equal(t, existing.Result.CreatedBy, result.Result.CreatedBy)
			require.True(t, existing.Result.Updated.UTC().Before(result.Result.Updated.UTC()))
			require.Equal(t, existing.Result.UpdatedBy, result.Result.UpdatedBy)
		})

	testScenario(t, "When an admin tries to update a library panel with model only, it should change model successfully and return correct result",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var existing libraryPanelResult
			err := json.Unmarshal(response.Body(), &existing)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				Model: []byte(`{ "name": "New Model Name" }`),
			}
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": existing.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 200, response.Status())
			var result libraryPanelResult
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, existing.Result.ID, result.Result.ID)
			require.Equal(t, existing.Result.OrgID, result.Result.OrgID)
			require.Equal(t, int64(1), result.Result.FolderID)
			require.Equal(t, existing.Result.Name, result.Result.Name)
			require.Equal(t, "New Model Name", result.Result.Model["name"])
			require.Equal(t, existing.Result.Created.UTC().Format(time.RFC3339), result.Result.Created.UTC().Format(time.RFC3339))
			require.Equal(t, existing.Result.CreatedBy, result.Result.CreatedBy)
			require.True(t, existing.Result.Updated.UTC().Before(result.Result.Updated.UTC()))
			require.Equal(t, existing.Result.UpdatedBy, result.Result.UpdatedBy)
		})

	testScenario(t, "When another admin tries to update a library panel, it should change UpdatedBy successfully and return correct result",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var existing libraryPanelResult
			err := json.Unmarshal(response.Body(), &existing)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{}
			sc.reqContext.UserId = 2
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": existing.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 200, response.Status())
			var result libraryPanelResult
			err = json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)
			require.Equal(t, existing.Result.ID, result.Result.ID)
			require.Equal(t, existing.Result.OrgID, result.Result.OrgID)
			require.Equal(t, existing.Result.FolderID, result.Result.FolderID)
			require.Equal(t, existing.Result.Name, result.Result.Name)
			require.Equal(t, existing.Result.Model["name"], result.Result.Model["name"])
			require.Equal(t, existing.Result.Created.UTC().Format(time.RFC3339), result.Result.Created.UTC().Format(time.RFC3339))
			require.Equal(t, existing.Result.CreatedBy, result.Result.CreatedBy)
			require.True(t, existing.Result.Updated.UTC().Before(result.Result.Updated.UTC()))
			require.Equal(t, int64(2), result.Result.UpdatedBy)
		})

	testScenario(t, "When an admin tries to update a library panel with a name that already exists, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Existing")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			command = getCreateCommand(1, "Text - Library Panel")
			response = sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				Name: "Existing",
			}
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 400, response.Status())
		})

	testScenario(t, "When an admin tries to update a library panel with a folder where a library panel with the same name already exists, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(2, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			command = getCreateCommand(1, "Text - Library Panel")
			response = sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				FolderID: 2,
			}
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 400, response.Status())
		})

	testScenario(t, "When an admin tries to update a library panel in another org, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreateCommand(1, "Text - Library Panel")
			response := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 200, response.Status())

			var result libraryPanelResult
			err := json.Unmarshal(response.Body(), &result)
			require.NoError(t, err)

			cmd := updateLibraryPanelCommand{
				FolderID: 2,
			}
			sc.reqContext.OrgId = 2
			sc.reqContext.ReplaceAllParams(map[string]string{":uid": result.Result.UID})
			response = sc.service.updateHandler(sc.reqContext, cmd)
			require.Equal(t, 404, response.Status())
		})
}

type libraryPanel struct {
	ID        int64                  `json:"ID"`
	OrgID     int64                  `json:"OrgID"`
	FolderID  int64                  `json:"FolderID"`
	UID       string                 `json:"UID"`
	Name      string                 `json:"Name"`
	Model     map[string]interface{} `json:"Model"`
	Created   time.Time              `json:"Created"`
	Updated   time.Time              `json:"Updated"`
	CreatedBy int64                  `json:"CreatedBy"`
	UpdatedBy int64                  `json:"UpdatedBy"`
}

type libraryPanelResult struct {
	Result libraryPanel `json:"result"`
}

type libraryPanelsResult struct {
	Result []libraryPanel `json:"result"`
}

func overrideLibraryPanelServiceInRegistry(cfg *setting.Cfg) LibraryPanelService {
	lps := LibraryPanelService{
		SQLStore: nil,
		Cfg:      cfg,
	}

	overrideServiceFunc := func(d registry.Descriptor) (*registry.Descriptor, bool) {
		descriptor := registry.Descriptor{
			Name:         "LibraryPanelService",
			Instance:     &lps,
			InitPriority: 0,
		}

		return &descriptor, true
	}

	registry.RegisterOverride(overrideServiceFunc)

	return lps
}

func getCreateCommand(folderID int64, name string) createLibraryPanelCommand {
	command := createLibraryPanelCommand{
		FolderID: folderID,
		Name:     name,
		Model: []byte(`
			{
			  "datasource": "${DS_GDEV-TESTDATA}",
			  "id": 1,
			  "name": "Text - Library Panel",
			  "type": "text"
			}
		`),
	}

	return command
}

type scenarioContext struct {
	ctx        *macaron.Context
	service    *LibraryPanelService
	reqContext *models.ReqContext
	user       models.SignedInUser
}

// testScenario is a wrapper around t.Run performing common setup for library panel tests.
// It takes your real test function as a callback.
func testScenario(t *testing.T, desc string, fn func(t *testing.T, sc scenarioContext)) {
	t.Helper()

	t.Run(desc, func(t *testing.T) {
		t.Cleanup(registry.ClearOverrides)

		ctx := macaron.Context{}
		orgID := int64(1)
		role := models.ROLE_ADMIN

		cfg := setting.NewCfg()
		// Everything in this service is behind the feature toggle "panelLibrary"
		cfg.FeatureToggles = map[string]bool{"panelLibrary": true}
		// Because the LibraryPanelService is behind a feature toggle, we need to override the service in the registry
		// with a Cfg that contains the feature toggle so migrations are run properly
		service := overrideLibraryPanelServiceInRegistry(cfg)

		// We need to assign SQLStore after the override and migrations are done
		sqlStore := sqlstore.InitTestDB(t)
		service.SQLStore = sqlStore

		user := models.SignedInUser{
			UserId:     1,
			OrgId:      orgID,
			OrgRole:    role,
			LastSeenAt: time.Now(),
		}
		sc := scenarioContext{
			user:    user,
			ctx:     &ctx,
			service: &service,
			reqContext: &models.ReqContext{
				Context:      &ctx,
				SignedInUser: &user,
			},
		}
		fn(t, sc)
	})
}
