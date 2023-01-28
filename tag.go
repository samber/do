package do

import (
	"fmt"
	"reflect"
	"unsafe"
)

const tagName = "do"

func InjectTag[T any](i *Injector, service T) (T, error) {
	_i := getInjectorOrDefault(i)
	return populateServiceByTags(_i, service)
}

func MustInjectTag[T any](i *Injector, service T) T {
	v, err := InjectTag(i, service)
	must(err)
	return v
}

func populateServiceByTags[T any](i *Injector, service T) (T, error) {
	// get the underlying value
	value := reflect.ValueOf(service)

	// return error if provided service is not a pointer
	if value.Kind() != reflect.Ptr {
		return empty[T](), fmt.Errorf("DI: expected a pointer")
	}

	// look for the underlying value type
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// check if underlying value is a struct
	kind := value.Kind()
	if kind != reflect.Struct {
		return service, nil
	}

	tYpe := value.Type()

	// for each field, we try to inject value
	for j := 0; j < tYpe.NumField(); j++ {
		field := tYpe.Field(j)

		tag, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		serviceName := tag
		if len(tag) == 0 {
			serviceName = getServiceNameFromStructField(&field)
		}

		// fmt.Println("lookup service:", serviceName)

		serviceAny, ok := i.get(serviceName)
		if !ok {
			return empty[T](), i.serviceNotFound(serviceName)
		}

		service, ok := serviceAny.(anyService)
		if !ok {
			return empty[T](), i.serviceNotFound(serviceName)
		}

		instance, err := service.getInstanceAny(i)
		if err != nil {
			return empty[T](), err
		}

		newValue := reflect.ValueOf(instance)
		if newValue.Type() != field.Type {
			return empty[T](), fmt.Errorf("DI: type mismatch. Expected '%s', got '%s'", newValue.Type().String(), field.Type.String())
		}

		// https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields
		reflect.NewAt(field.Type, unsafe.Pointer(value.Field(j).UnsafeAddr())).
			Elem().
			Set(newValue)
	}

	return service, nil
}
