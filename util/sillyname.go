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
	"dog", "puppy", "turtle", "rabbit", "parrot", "cat", "kitten", "goldfish",
	"mouse", "hamster", "fish", "cow", "rabbit", "duck", "shrimp", "pig",
	"goat", "crab", "deer", "bee", "sheep", "fish", "turkey", "dove",
	"chicken", "horse", "squirrel", "dog", "chimpanzee", "ox", "lion", "panda",
	"walrus", "otter", "mouse", "kangaroo", "goat", "horse", "monkey", "cow",
	"koala", "mole", "elephant", "leopard", "hippopotamus", "giraffe", "fox",
	"coyote", "hedgehong", "sheep", "deer",
}

var colors = []string{
	"amaranth", "amber", "amethyst", "apricot", "aqua", "aquamarine", "azure",
	"beige", "black", "blue", "blush", "bronze", "brown", "chocolate",
	"coffee", "copper", "coral", "crimson", "cyan", "emerald", "fuchsia",
	"gold", "gray", "green", "harlequin", "indigo", "ivory", "jade",
	"lavender", "lime", "magenta", "maroon", "moccasin", "olive", "orange",
	"peach", "pink", "plum", "purple", "red", "rose", "salmon", "sapphire",
	"scarlet", "silver", "tan", "teal", "tomato", "turquoise", "violet",
	"white", "yellow",
}

func r(r *rand.Rand, l []string) string {
	return l[r.Intn(len(l)-1)]
}

func NewUserName(s *rand.Rand) string {
	title := cases.Title(language.English)
	return title.String(r(s, adjectives)) + " " + title.String(r(s, animals))
}

func NewRoomName(s *rand.Rand) string {
	return r(s, adjectives) + "-" + r(s, colors) + "-" + r(s, animals)
}
