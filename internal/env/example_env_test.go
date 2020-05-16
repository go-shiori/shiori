package env

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

var (
	Prefix string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	Prefix = fmt.Sprintf("%08d_", rand.Uint32())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ExampleGetEnvString() {

	key := Prefix + "TEST00"
	must(os.Setenv(key, "BAR"))
	fmt.Println(GetEnvString(key, "FOO"))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvString(key, "FOO"))

	// Output:
	// BAR
	// FOO
}

func ExampleGetEnvStringList() {

	key := Prefix + "TEST10"
	must(os.Setenv(key, "FOO, BAR,BAZ"))
	fmt.Println(GetEnvStringList(key, []string{"FOO"}))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvStringList(key, []string{"FOO"}))

	key = Prefix + "TEST11"
	must(os.Setenv(key, "FOO : BAR:BAZ"))
	fmt.Println(GetEnvStringList(key, []string{"FOO"}))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvStringList(key, []string{"FOO"}))

	// Output:
	// [FOO BAR BAZ]
	// [FOO]
	// [FOO BAR BAZ]
	// [FOO]
}

func ExampleGetEnvInt64() {

	key := Prefix + "TEST20"
	must(os.Setenv(key, "43"))
	fmt.Println(GetEnvInt64(key, 42))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvInt64(key, 42))

	key = Prefix + "TEST21"
	must(os.Setenv(key, "dumb"))
	fmt.Println(GetEnvInt64(key, 42))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvInt64(key, 42))

	// Output:
	// 43
	// 42
	// 42
	// 42
}

func ExampleGetEnvBool() {

	key := Prefix + "TEST30"
	must(os.Setenv(key, "Y"))
	fmt.Println(GetEnvBool(key, false))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvBool(key, false))

	key = Prefix + "TEST31"
	must(os.Setenv(key, "False"))
	fmt.Println(GetEnvBool(key, true))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvBool(key, true))

	key = Prefix + "TEST32"
	must(os.Setenv(key, "dumb"))
	fmt.Println(GetEnvBool(key, false))
	must(os.Unsetenv(key))
	fmt.Println(GetEnvBool(key, false))

	// Output:
	// true
	// false
	// false
	// true
	// false
	// false
}
