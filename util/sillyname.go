package util

import (
	"math/rand"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func r(r *rand.Rand, l []string) string {
	return l[r.Intn(len(l)-1)]
}

func NewName(s *rand.Rand) string {
	return cases.Title(language.English).String(r(s, adjectives) + " " + r(s, animals))
}
