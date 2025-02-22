package notorm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitrevance/notorm"
)

type CallbackScanner struct {
	cb func(...any) error
}

func (s *CallbackScanner) Scan(fields ...any) error {
	return s.cb(fields...)
}

type CallbackColumnAwareScanner struct {
	CallbackScanner
	cols func() ([]string, error)
}

func (s *CallbackColumnAwareScanner) Columns() ([]string, error) {
	return s.cols()
}

func TestScanner(t *testing.T) {
	type InnerObj struct {
		F3      float32 `column:"field_3"`
		Field_3 int
	}
	type InnerObjByPtr struct {
		F4 float32
	}
	type Obj struct {
		InnerObj
		*InnerObjByPtr
		F1 int
		F2 string
	}

	v := &Obj{InnerObjByPtr: &InnerObjByPtr{}}

	assert.NoError(t, notorm.Scan(v, &CallbackColumnAwareScanner{
		cols: func() ([]string, error) {
			return []string{"F1", "F2", "field_3", "F4"}, nil
		},
		CallbackScanner: CallbackScanner{
			cb: func(a ...any) error {
				assert.Equal(t, &v.F1, a[0])
				assert.Equal(t, &v.F2, a[1])
				assert.Equal(t, &v.F3, a[2])
				assert.Equal(t, &v.F4, a[2])
				return nil
			},
		},
	}))
}

func TestScannerError(t *testing.T) {
	type InnerObj struct {
		F3      float32 `column:"field_3"`
		Field_3 int
		F4      string
	}
	type InnerObjByPtr struct {
		F4 float32
	}
	type Obj struct {
		InnerObj
		*InnerObjByPtr
		F1 int
		F2 string
	}

	v := &Obj{InnerObjByPtr: &InnerObjByPtr{}}

	assert.Error(t, notorm.Scan(v, &CallbackColumnAwareScanner{
		cols: func() ([]string, error) {
			return []string{"f1", "f2", "field_3", "F4"}, nil
		},
		CallbackScanner: CallbackScanner{
			cb: func(a ...any) error {
				assert.Fail(t, "should have failed earlier")
				return nil
			},
		},
	}))
}

func TestSchema(t *testing.T) {
	type Data struct {
		Id    int `column:"data_id"`
		Name  string
		Email string
		Count int
	}

	schema := notorm.SchemaOf[Data]()

	assert.Equal(t, "data_id", schema.NameOf(&schema.C.Id))
	assert.Equal(t, "Name", schema.NameOf(&schema.C.Name))
	assert.Equal(t, "Email", schema.NameOf(&schema.C.Email))
	assert.Equal(t, "Count", schema.NameOf(&schema.C.Count))
}

func TestSchemaPrefix(t *testing.T) {
	type Data struct {
		Id    int `column:"data_id"`
		Name  string
		Email string
		Count int
	}
	type ComplexData struct {
		Data
		Data2   Data
		Data3   Data `prefix:"admin_"`
		Content string
	}

	schema := notorm.SchemaOf[ComplexData]()

	assert.Equal(t, "data_id", schema.NameOf(&schema.C.Id))
	assert.Equal(t, "Name", schema.NameOf(&schema.C.Name))
	assert.Equal(t, "Email", schema.NameOf(&schema.C.Email))
	assert.Equal(t, "Count", schema.NameOf(&schema.C.Count))
	assert.Equal(t, "Content", schema.NameOf(&schema.C.Content))

	assert.Equal(t, "Data2.Name", schema.NameOf(&schema.C.Data2.Name))
	assert.Equal(t, "admin_Name", schema.NameOf(&schema.C.Data3.Name))
}
