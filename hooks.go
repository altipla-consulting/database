package database

import (
	"reflect"
	"time"

	"github.com/juju/errors"
)

// BeforeCreater should be implemented in models that want to run custom
// logic before creating a new model in the DB.
type BeforeCreater interface {
	BeforeCreate() error
}

func runBeforeCreateHook(model reflect.Value) error {
	t, ok := model.Interface().(BeforeCreater)
	if ok {
		if err := t.BeforeCreate(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()
	now := time.Now()

	// Change create date if the field is present
	createdAt := modelElem.FieldByName("CreatedAt")
	if createdAt.IsValid() {
		createdAt.Set(reflect.ValueOf(now))
	}

	// Change update date if the field is present
	updatedAt := modelElem.FieldByName("UpdatedAt")
	if updatedAt.IsValid() {
		updatedAt.Set(reflect.ValueOf(now))
	}

	t, ok = modelElem.Interface().(BeforeCreater)
	if ok {
		if err := t.BeforeCreate(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// AfterCreater should be implemented in models that want to run custom
// logic after creating a new model in the DB.
type AfterCreater interface {
	AfterCreate() error
}

func runAfterCreateHook(model reflect.Value) error {
	t, ok := model.Interface().(AfterCreater)
	if ok {
		if err := t.AfterCreate(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()
	t, ok = modelElem.Interface().(AfterCreater)
	if ok {
		if err := t.AfterCreate(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// BeforeUpdater should be implemented in models that want to run custom
// logic before updating a stored model.
type BeforeUpdater interface {
	BeforeUpdate() error
}

func runBeforeUpdateHook(model reflect.Value) error {
	t, ok := model.Interface().(BeforeUpdater)
	if ok {
		if err := t.BeforeUpdate(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()

	// Change update date if the field is present
	updatedAt := modelElem.FieldByName("UpdatedAt")
	if updatedAt.IsValid() {
		updatedAt.Set(reflect.ValueOf(time.Now()))
	}

	t, ok = modelElem.Interface().(BeforeUpdater)
	if ok {
		if err := t.BeforeUpdate(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// AfterUpdater should be implemented in models that want to run custom
// logic after updating a stored model.
type AfterUpdater interface {
	AfterUpdate() error
}

func runAfterUpdateHook(model reflect.Value) error {
	t, ok := model.Interface().(AfterUpdater)
	if ok {
		if err := t.AfterUpdate(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()
	t, ok = modelElem.Interface().(AfterUpdater)
	if ok {
		if err := t.AfterUpdate(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// BeforeSaver should be implemented in models that want to run custom
// logic before saving a model.
type BeforeSaver interface {
	BeforeSave() error
}

func runBeforeSaveHook(model reflect.Value) error {
	t, ok := model.Interface().(BeforeSaver)
	if ok {
		if err := t.BeforeSave(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()
	t, ok = modelElem.Interface().(BeforeSaver)
	if ok {
		if err := t.BeforeSave(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// AfterSaver should be implemented in models that want to run custom
// logic after saving a model.
type AfterSaver interface {
	AfterSave() error
}

func runAfterSaveHook(model reflect.Value) error {
	t, ok := model.Interface().(AfterSaver)
	if ok {
		if err := t.AfterSave(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()
	t, ok = modelElem.Interface().(AfterSaver)
	if ok {
		if err := t.AfterSave(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// AfterFinder should be implemented in models that want to run custom
// logic after fetching a model from DB.
type AfterFinder interface {
	AfterFind() error
}

func runAfterFindHook(model reflect.Value) error {
	t, ok := model.Interface().(AfterFinder)
	if ok {
		if err := t.AfterFind(); err != nil {
			return errors.Trace(err)
		}
	}

	modelElem := model.Elem()
	t, ok = modelElem.Interface().(AfterFinder)
	if ok {
		if err := t.AfterFind(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
