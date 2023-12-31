package entities

import (
	"github.com/icloudeng/platform-installer/internal/database"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	tfjson "github.com/hashicorp/terraform-json"
)

type StateType map[string]*tfjson.StateResource

type ResourcesState struct {
	gorm.Model
	Ref         string `gorm:"index"`
	State       datatypes.JSON
	Credentials datatypes.JSON
	JobID       uint
	Job         Job `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type ResourcesStateRepository struct{}

func (ResourcesStateRepository) GetByRef(ref string) *ResourcesState {
	var object ResourcesState

	database.Conn.Joins("Job").Where(&ResourcesState{
		Ref: ref,
	}).Last(&object)

	if object.ID == 0 {
		return nil
	}

	return &object
}

func (ResourcesStateRepository) Get(ID uint) *ResourcesState {
	object := &ResourcesState{}

	database.Conn.Last(object, ID)

	if object.ID == 0 {
		return nil
	}

	return object
}

func (ResourcesStateRepository) Create(res *ResourcesState) {
	database.Conn.Create(res)
}

func (ResourcesStateRepository) UpdateOrCreate(res *ResourcesState) {
	database.Conn.Save(res)
}

func (ResourcesStateRepository) Delete(ID uint) {
	database.Conn.Delete(&ResourcesState{}, ID)
}

func init() {
	database.Conn.AutoMigrate(&ResourcesState{})
}
