package main

import "fmt"


type Cat struct {

}

func (c Cat) Speak() {
	fmt.Print("meow")
}

type Dog struct {}

func (d Dog) Speak() {
	fmt.Print("woof")
}

type SpeakerPrefixer struct {
	Speaker
}
func (sp SpeakerPrefixer) Speak() {
	fmt.Println("Prefix:")
	sp.Speaker.Speak()
}


type Husky struct {
	Speaker
}

type Speaker interface {
	Speak()
}

func main() {
	h := Husky{SpeakerPrefixer{Dog{}}}
	h.Speak() // equivalent to h.Dog.Speak()
}
