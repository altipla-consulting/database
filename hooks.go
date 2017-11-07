package database

type ModelTrackingAfterGetHooker interface {
	ModelTrackingAfterGet(props []*Property) error
}

type ModelTrackingAfterPutHooker interface {
	ModelTrackingAfterPut(props []*Property) error
}
