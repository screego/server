package util

import (
	"math/rand"
	"strings"
)

var adjectives = []string{
	"able", "adaptive", "adventurous", "affable", "agreeable", "ambitious",
	"amiable", "amusing", "balanced", "brave", "bright", "calm", "capable",
	"charming", "clever", "compassionate", "considerate", "courageous",
	"creative", "decisive", "determined", "discreet", "dynamic",
	"enthusiastic", "exuberant", "faithful", "fearless", "friendly", "funny",
	"generous", "gentle", "good", "honest", "humorous", "independent",
	"intelligent", "intuitive", "kind", "loving", "loyal", "modest", "nice",
	"optimistic", "patient", "pioneering", "polite", "powerful", "reliable",
	"resourceful", "sensible", "sincere", "thoughtful", "tough", "versatile",
}
var animals = []string{
	"Dog", "Puppy", "Turtle", "Rabbit", "Parrot", "Cat", "Kitten", "Goldfish",
	"Mouse", "Hamster", "Fish", "Cow", "Rabbit", "Duck", "Shrimp", "Pig",
	"Goat", "Crab", "Deer", "Bee", "Sheep", "Fish", "Turkey", "Dove",
	"Chicken", "Horse", "Squirrel", "Dog", "Chimpanzee", "Ox", "Lion", "Panda",
	"Walrus", "Otter", "Mouse", "Kangaroo", "Goat", "Horse", "Monkey", "Cow",
	"Koala", "Mole", "Elephant", "Leopard", "Hippopotamus", "Giraffe", "Fox",
	"Coyote", "Hedgehong", "Sheep", "Deer",
}

func r(l []string) string {
	return l[rand.Intn(len(l)-1)]
}

func NewName() string {
	return strings.Title(r(adjectives) + " " + r(animals))
}
