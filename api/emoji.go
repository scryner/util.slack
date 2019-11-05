package api

import (
	"math/rand"
)

var (
	emoji    []string
	lenEmoji int
)

func init() {
	emoji = []string{
		":blush:",
		":relaxed:",
		":heart_eyes:",
		":relieved:",
		":grinning:",
		":kissing_smiling_eyes:",
		":open_mouth:",
		":smiling_imp:",
		":no_mouth:",
		":no_mouth:",
		":blue_heart:",
		":heart:",
		":heartpulse:",
		":revolving_hearts:",
		":sparkling_heart:",
		":star:",
		":musical_note:",
		":angel:",
		":smiley_cat:",
		":heart_eyes_cat:",
		":smirk_cat:",
		":smile:",
		":laughing:",
		":smiley:",
		":wink:",
		":yellow_heart:",
		":purple_heart:",
		":green_heart:",
		":heartbeat:",
		":two_hearts:",
		":sparkles:",
		":star2:",
		":smile_cat:",
		":sunny:",
		":cloud:",
		":ocean:",
		":dog:",
		":hamster:",
		":wolf:",
		":tiger:",
		":bear:",
		":monkey:",
		":sheep:",
		":panda_face:",
		":bird:",
		":turtle:",
		":beetle:",
		":octopus:",
		":fish:",
		":whale2:",
		":cherry_blossom:",
		":four_leaf_clover:",
		":maple_leaf:",
		":deciduous_tree:",
		":mushroom:",
		":full_moon:",
		":crescent_moon:",
		":crescent_moon:",
		":cat:",
		":rabbit:",
		":frog:",
		":koala:",
		":monkey_face:",
		":elephant:",
		":tropical_fish:",
		":snail:",
		":whale:",
		":dolphin:",
		":dragon_face:",
		":poodle:",
		":bouquet:",
		":tulip:",
		":rose:",
		":hibiscus:",
		":leaves:",
		":herb:",
		":cactus:",
		":evergreen_tree:",
		":shell:",
		":earth_asia:",
		":bamboo:",
		":jack_o_lantern:",
		":ghost:",
	}

	lenEmoji = len(emoji)
}

func RandEmoji() string {
	i := rand.Int() % lenEmoji
	return emoji[i]
}
