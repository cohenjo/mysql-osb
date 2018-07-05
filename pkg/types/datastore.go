package types

type ObjectType string

const (
	Instance ObjectType = "instance"
	Binding             = "binding"
)

type BranchType string

const (
	ProductionDB  BranchType = "production-database"
	StagingDB     BranchType = "staging-database"
	DevelopmentDB BranchType = "development-database"
	TestBed       BranchType = "TEST_BED"
	NoBranch      BranchType = "no_branch"
)

type DataStore interface {
	Set(typePar ObjectType, key string, branch BranchType, version string, value interface{}) (err error)
	Get(typePar ObjectType, key string, branch BranchType, version string) (retVal interface{}, err error)
	Remove(typePar ObjectType, key string, branch BranchType, version string) (err error)
	RemoveKey(typePar ObjectType, key string) (err error)
	RemoveType(typePar ObjectType)
}
