package helper

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Insert data by batch can increase database performance drastically, this class aim to make batch-save easier,
// It takes care the database operation for specified `slotType`, records got saved into database whenever cache hits
// The `size` limit, remember to call the `Close` method to save the last batch
type BatchSave struct {
	slotType reflect.Type
	// slots can not be []interface{}, because gorm wouldn't take it
	// I'm guessing the reason is the type information lost when converted to interface{}
	slots         reflect.Value
	db            *gorm.DB
	current       int
	size          int
	valueIndex    map[string]int
	hasPrimarykey bool
}

func NewBatchSave(db *gorm.DB, slotType reflect.Type, size int) (*BatchSave, error) {
	if slotType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("slotType must be a pointer")
	}

	return &BatchSave{
		slotType:      slotType,
		slots:         reflect.MakeSlice(reflect.SliceOf(slotType), size, size),
		db:            db,
		size:          size,
		valueIndex:    make(map[string]int),
		hasPrimarykey: true,
	}, nil
}

func (c *BatchSave) Add(slot interface{}) error {
	// type checking
	if reflect.TypeOf(slot) != c.slotType {
		return fmt.Errorf("sub cache type mismatched")
	}
	if reflect.ValueOf(slot).Kind() != reflect.Ptr {
		return fmt.Errorf("slot is not a pointer")
	}
	// deduplication
	key := getPrimaryKeyValue(slot)
	c.hasPrimarykey = hasPrimaryKey(slot)
	
	if key != "" {
		if index, ok := c.valueIndex[key]; !ok {
			c.valueIndex[key] = c.current
		} else {
			c.slots.Index(index).Set(reflect.ValueOf(slot))
			return nil
		}
	}
	c.slots.Index(c.current).Set(reflect.ValueOf(slot))
	c.current++
	// flush out into database if max outed
	if c.current == c.size {
		return c.Flush()
	}
	return nil
}

func (c *BatchSave) Flush() error {
	if c.hasPrimarykey {
		err := c.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(c.slots.Slice(0, c.current).Interface()).Error
		if err != nil {
			return err
		}
	} else {
		err := c.db.Create(c.slots.Slice(0, c.current).Interface()).Error
		if err != nil {
			return err
		}
	}

	c.current = 0
	c.valueIndex = make(map[string]int)
	return nil
}

func (c *BatchSave) Close() error {
	if c.current > 0 {
		return c.Flush()
	}
	return nil
}

func isPrimaryKey(f reflect.StructField, v reflect.Value) (string, bool) {
	tag := strings.TrimSpace(f.Tag.Get("gorm"))
	if strings.HasPrefix(strings.ToLower(tag), "primarykey") {
		return fmt.Sprintf("%v", v.Interface()), true
	}
	return "", false
}

func hasPrimaryKey(iface interface{}) bool {
	ifv := reflect.ValueOf(iface)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		v := ifv.Field(i)
		switch v.Kind() {
		case reflect.Struct:
			ok := hasPrimaryKey(v.Interface())
			if ok {
				return true
			}
		default:
			if _, ok := isPrimaryKey(ifv.Type().Field(i), v); ok {
				return true
			}
		}
	}

	return false
}

func getPrimaryKeyValue(iface interface{}) string {
	var ss []string
	ifv := reflect.ValueOf(iface)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		v := ifv.Field(i)
		switch v.Kind() {
		case reflect.Struct:
			s := getPrimaryKeyValue(v.Interface())
			if s != "" {
				ss = append(ss, s)
			}
		default:
			if s, ok := isPrimaryKey(ifv.Type().Field(i), v); ok && s != "" {
				ss = append(ss, s)
			}
		}
	}
	return strings.Join(ss, ":")
}
