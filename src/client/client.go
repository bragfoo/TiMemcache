package main

import (
	"log"

	"./memcache"
)

func main() {
	client := memcache.New("127.0.0.1:11212")
	// log.Println("client", client.Timeout, client.MaxIdleConns, client.Selector, client.Freeconn)

	// set
	log.Println("set operta 1")
	setErr1 := client.Set(&memcache.Item{Key: "foo1", Value: []byte("bar")})
	log.Println(setErr1)

	log.Println("set operta 2")
	setErr2 := client.Set(&memcache.Item{Key: "foo2", Value: []byte("bar")})
	log.Println(setErr2)

	log.Println("set operta 3")
	setErr3 := client.Set(&memcache.Item{Key: "foo3", Value: []byte("bar")})
	log.Println(setErr3)

	log.Println("set operta 4")
	setErr4 := client.Set(&memcache.Item{Key: "foo4", Value: []byte("bar")})
	log.Println(setErr4)

	// add
	log.Println("add operta")
	addErr := client.Add(&memcache.Item{Key: "foo", Value: []byte("bluegogo")})
	log.Println(addErr)

	// get
	log.Println("get operta")
	getVal, getErr := client.Get("foo")
	if getErr != nil {
		log.Println("getErr:", getErr)
	} else {
		log.Println(string(getVal.Value))
	}

	// replace
	// log.Println("replace operta")
	// client.Replace(&memcache.Item{Key: "foo", Value: []byte("mobike")})
	// replaceVal, replaceErr := client.Get("foo")
	// if replaceErr != nil {
	// 	log.Println(replaceErr)
	// } else {
	// 	log.Println(replaceVal)
	// }

	// // delete
	// log.Println("delete operta")
	// delErr := client.Delete("foo")
	// if delErr != nil {
	// 	log.Println("Delete failed:", delErr)
	// }

	// // get
	// log.Println("get operta")
	// dgetVal, dgetErr := client.Get("foo")
	// if dgetVal != nil {
	// 	log.Println(dgetErr)
	// } else {
	// 	log.Println(dgetVal)
	// }

	// incr
	// client.Set(&memcache.Item{Key: "hoo", Value: []byte("1")})
	// item, _ = client.Get("hoo")
	// fmt.Println(item.Key, string(item.Value), item.Flags, item.Expiration)
	// incrVal, _ := client.Increment("hoo", 5)
	// fmt.Println(incrVal)
	// // decrby
	// decrVal, _ := client.Decrement("hoo", 4)
	// fmt.Println(decrVal)
	// // get multi
	// multiVal, getsErr := client.GetMulti([]string{"foo", "hoo"})
	// if getsErr != nil {
	// 	fmt.Println(getsErr)
	// } else {
	// 	fmt.Println(multiVal)
	// }
	// // flush all
	// client.FlushAll()
	// get multi
	// multiVal, getsErr = client.GetMulti([]string{"foo", "hoo"})
	// if getsErr != nil {
	// 	fmt.Println(getsErr)
	// } else {
	// 	fmt.Println(multiVal)
	// }
	// // delete all
	// client.DeleteAll()
	// // get multi
	// multiVal, _ = client.GetMulti([]string{"foo", "hoo"})
	// fmt.Println(multiVal)
}
