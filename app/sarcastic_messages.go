package main

import (
	"crypto/rand"
	"math/big"
)

// Map seconds spent to random sarcastic messages.
var SarcasticTimes = map[int64][]string{
	// Under 5 seconds.
	int64(0): {
		"Hmm, that was quick. I guess Minesweeper isn't much of a challenge for you.",
		"Wow! Are you sure you even played Minesweeper? Or did you just click randomly?",
		"Congratulations on completing Minesweeper at warp speed. I hope you found it as stimulating as watching paint dry.",
		"You finished Minesweeper in no time. I suppose your superhuman speed is better suited for something meaningful. Like staring into space.",
		"That was lightning-fast! You must have a natural talent for clicking on squares. Too bad it won't get you anywhere in life.",
	},
	// 5 seconds or more.
	int64(5): {
		"Impressive! You finished a Minesweeper board so quickly. Are you always in such a hurry? Maybe you should slow down and savor your inevitable defeat in other endeavors too.",
		"Wow, look at you finishing Minesweeper in a flash. Your speed is truly remarkable. Or maybe it's just a sign that you have nothing better to do.",
		"That was lightning-fast! I'm amazed at your ability to click randomly and still come out victorious. You should consider joining the Minesweeper Olympics.",
		"So fast! Did you even have time to blink? Your Minesweeper prowess is unmatched. Too bad it's utterly meaningless.",
		"Over 5 seconds? That's barely enough time for a single thought. But hey, congratulations on your groundbreaking Minesweeper achievement.",
	},
	// 10 seconds or more.
	int64(10): {
		"Impressive! You finished a Minesweeper board so quickly. Are you always in such a hurry? Maybe you should slow down and savor your inevitable defeat in other endeavors too.",
		"Wow, look at you finishing Minesweeper in a flash. Your speed is truly remarkable. Or maybe it's just a sign that you have nothing better to do.",
		"That was lightning-fast! I'm amazed at your ability to click randomly and still come out victorious. You should consider joining the Minesweeper Olympics.",
		"So fast! Did you even have time to blink? Your Minesweeper prowess is unmatched. Too bad it's utterly meaningless.",
		"Ten seconds? That's barely enough time for a single thought. But hey, congratulations on your groundbreaking Minesweeper achievement.",
	},
	// 15 seconds or more.
	int64(15): {
		"Congratulations! You finished Minesweeper fast. I guess even a broken clock finds a mine sometimes.",
		"Fifteen seconds? You're practically a Minesweeper prodigy. Maybe you should consider joining the Olympic team for speed Minesweeping.",
		"You finished Minesweeper in just 15 seconds? That's faster than the time it takes for your computer to crash. Impressive!",
		"Fifteen seconds? If only you could apply that kind of efficiency to something that actually matters.",
		"Fifteen seconds? That's quicker than the lifespan of a fruit fly. Don't worry, your Minesweeper achievement won't live long in anyone's memory either.",
	},
	// 20 seconds or more.
	int64(20): {
		"Oh, look at you, finishing Minesweeper in record time. I bet your therapist wishes you had that kind of speed during your sessions.",
		"Twenty seconds? Are you sure you didn't use a time machine to cheat? You're clearly too skilled for this world.",
		"Twenty seconds? I'm pretty sure you just set a new Minesweeper world record. Oh wait, nobody cares.",
		"Twenty seconds? That's faster than it takes for your disappointment to sink in. Congrats on your fleeting success.",
		"Twenty seconds? I'm starting to suspect you're a Minesweeper wizard trapped in the body of an underachiever.",
	},
	// 30 seconds or more.
	int64(30): {
		"Wow, you completed Minesweeper so fast. I'm impressed! Not with your skills, but with how little you contribute to society.",
		"Thirty seconds? That's faster than it takes for people to realize they've made a mistake talking to you.",
		"Thirty seconds? You're like a ninja of Minesweeper. Silent, swift, and utterly inconsequential.",
		"Thirty seconds? Your ability to click randomly and survive never ceases to amaze me. Keep up the aimless work!",
		"Thirty seconds? I suppose that's one way to avoid dealing with your real-life problems.",
	},
	// 40 seconds or more.
	int64(40): {
		"You finished Minesweeper quickly. I hope you have a lot of free time on your hands because that's the only thing you're good at.",
		"Forty seconds? You're like the Usain Bolt of Minesweeper. Just without the fame, talent, or meaningful achievements.",
		"Forty seconds? I'm starting to suspect you're secretly a robot programmed for Minesweeper domination.",
		"Forty seconds? That's shorter than your attention span. Good thing Minesweeper is right up your alley.",
		"Forty seconds? I guess miracles do happen. Even someone like you can finish Minesweeper in a reasonable time.",
	},
	// 50 seconds or more.
	int64(50): {
		"Congratulations! You finished Minesweeper almost as fast as it took for people to realize you're a disappointment.",
		"Fifty seconds? That's less time than it takes for people to regret inviting you to their parties.",
		"Fifty seconds? I bet you could finish Minesweeper blindfolded. Not that anyone would care.",
		"Fifty seconds? Keep up the incredible work. You're on your way to becoming the world champion of trivial accomplishments.",
		"Fifty seconds? You've certainly mastered the art of wasting time. Now, if only you could do something productive with it.",
	},
	// 1 minute or more.
	int64(60): {
		"Quick fingers, huh? Shame your speed doesn't translate to anything useful. Keep chasing that empty accomplishment.",
		"Sixty seconds? You're a real speed demon. I'm sure the world is in awe of your Minesweeper prowess.",
		"Sixty seconds? You're so fast, you make the Flash look like he's standing still. Well, not really.",
		"Sixty seconds? Not bad, but I'm sure you could've done better if you weren't so easily distracted by shiny objects.",
		"Sixty seconds? A true testament to your ability to excel at unimportant tasks. Bravo!",
	},
	// 1 minute and 10 seconds or more.
	int64(70): {
		"You finished Minesweeper fast, but it's still longer than your attention span. Have you tried coloring books instead?",
		"Seventy seconds? That's more than enough time for everyone to realize you're not as amazing as you think.",
		"Seventy seconds? I bet you could have solved world hunger in that time. Or maybe just eaten a sandwich.",
		"Seventy seconds? Keep up the great work. At this rate, you might actually accomplish something substantial in 10,000 years.",
		"Seventy seconds? You've reached a new level of mediocrity. Revel in your triumph!",
	},
	// 1 minute and a half or more.
	int64(90): {
		"Minesweeper, huh? Your skills are as impressive as a stale joke. Keep wasting your time; it's all you're good at.",
		"Ninety seconds? You're like a fine wine. Except not really, because you're about as enjoyable as vinegar.",
		"Ninety seconds? I've seen snails move faster. They should hire you to teach them the art of slowness.",
		"Ninety seconds? Congratulations on completing Minesweeper at a pace that even a sloth would find embarrassing.",
		"Ninety seconds? Your dedication to pointlessness is truly inspiring. Keep reaching for the stars!",
	},
	// 2 minutes or more.
	int64(60 * 2): {
		"You finished Minesweeper relatively quickly. I guess mediocrity is your thing. Congratulations on your unremarkable achievement.",
		"Two minutes? That's just enough time for people to forget you exist.",
		"Two minutes? You're a true Minesweeper wizard. If only there were a wizarding school for the profoundly unimpressive.",
		"Two minutes? Keep up the outstanding work. With your remarkable skills, you might just conquer the world of mediocrity.",
		"Two minutes? You're a shining example of how to excel at unimportant things. The world salutes you!",
	},
	// 5 minutes or more.
	int64(60 * 5): {
		"If Minesweeper is too challenging for you, I suggest trying Tic Tac Toe. It's a game more suitable for your level of incompetence.",
		"Five minutes? That's longer than the average attention span of a goldfish. Congratulations on your extraordinary focus.",
		"Five minutes? I'm surprised you didn't solve world peace in that time. Oh well, maybe next time.",
		"Five minutes? A truly magnificent display of your ability to prioritize the irrelevant. Keep it up!",
		"Five minutes? Your dedication to the art of futility is truly inspiring. Never stop reaching for the bottom!",
	},
	// 10 minutes or more.
	int64(60 * 10): {
		"Are you still playing Minesweeper? I thought I heard snails move faster. It's no surprise you're a champion in procrastination.",
		"Ten minutes? That's more time than it takes for your friends to find an excuse to leave when you're around.",
		"Ten minutes? You must have the patience of a saint. Or the lack of anything better to do.",
		"Ten minutes? Keep up the remarkable work. I'm sure you'll achieve greatness in the realm of insignificance.",
		"Ten minutes? Your unwavering commitment to trivial pursuits is truly awe-inspiring. Never stop chasing those meaningless victories!",
	},
	// 15 minutes or more.
	int64(60 * 15): {
		"You know, they say patience is a virtue. I guess Minesweeper is just too virtuous for you.",
		"Fifteen minutes? That's enough time for people to plan their escape when you start talking about your Minesweeper triumphs.",
		"Fifteen minutes? I'm starting to suspect you have a secret love affair with Minesweeper. It's okay, we won't judge.",
		"Fifteen minutes? Keep up the outstanding work. Your dedication to the inconsequential is truly unmatched!",
		"Fifteen minutes? Congratulations on your exceptional ability to waste time on activities that matter to no one!",
	},
	// 20 minutes or more.
	int64(60 * 20): {
		"Did you forget you were playing Minesweeper? It's no wonder your memory is as faulty as your gameplay.",
		"Twenty minutes? That's longer than your attention span during an important conversation.",
		"Twenty minutes? I hope you're taking notes because this Minesweeper expertise will surely come in handy someday. Not.",
		"Twenty minutes? Keep up the brilliant work. Your commitment to the unimportant is truly commendable!",
		"Twenty minutes? Your ability to excel at tasks of no consequence is truly remarkable. You're an inspiration to us all!",
	},
	// 30 minutes or more/
	int64(60 * 30): {
		"Thirty minutes? Seriously? I've seen glaciers move faster than you. Maybe you should try playing Minesweeper in geological time.",
		"Thirty minutes? You're really pushing the boundaries of human endurance with your Minesweeper marathons.",
		"Thirty minutes? That's enough time to write a novel about your incredible Minesweeper journey. Or not.",
		"Thirty minutes? Keep up the extraordinary work. Your dedication to the trivial knows no bounds!",
		"Thirty minutes? Your ability to dedicate copious amounts of time to the utterly insignificant is truly praiseworthy. Kudos!",
	},
	// 40 minutes or more.
	int64(60 * 40): {
		"You spent 40 minutes on Minesweeper? How precious. It must be nice to have nothing better to do with your life.",
		"Forty minutes? That's longer than your average nap time. Maybe you should consider a career as a professional Minesweeper sleeper.",
		"Forty minutes? I'm starting to think Minesweeper is your version of meditation. Just without the tranquility or enlightenment.",
		"Forty minutes? Keep up the incredible work. With your unparalleled dedication, you might just become the world's slowest Minesweeper champion!",
		"Forty minutes? Your ability to devote substantial portions of your life to meaningless endeavors is truly inspiring. Well done!",
	},
	// 50 minutes or more.
	int64(60 * 50): {
		"Congratulations on wasting 50 minutes of your life on Minesweeper. You must be the epitome of productivity.",
		"Fifty minutes? That's longer than it takes for a sloth to complete a marathon. Keep setting those ambitious goals!",
		"Fifty minutes? I hope you enjoyed every moment of your Minesweeper saga. Just kidding, nobody cares.",
		"Fifty minutes? That's enough time to question every life decision that led you to this moment of utter pointlessness.",
		"Fifty minutes? You've truly mastered the art of squandering time on the most inconsequential tasks. Keep up the magnificent work!",
	},
}

var SarcasticOneClickMessages = []string{
	"Wow, you're so lucky! You got a single click win! That board was almost as useless as you are!",
	"Well, well, well, aren't you just a prodigious prodigy of pointlessness? Bravo on your impeccable talent for wasting your own time.",
	"Congratulations, you've managed to reduce a game of strategy to a mindless game of chance. How utterly remarkable.",
	"Oh, splendid! I see you've perfected the art of mind-numbing luck. Perhaps next, you could teach us all how to stumble upon treasure without searching for it.",
	"My, my, what an extraordinary achievement! I must commend your skill in turning a mental exercise into a pitiful exercise in futility.",
	"One click, huh? You must have the intellectual prowess of a fungus to stumble upon the right square without even trying.",
	"Impressive, truly impressive! I had no idea someone could achieve mediocrity with such ease. Kudos to you, master of nothingness.",
	"How fortunate for you to discover the secret code to randomness! It takes a special kind of genius to excel in such meaningless endeavors.",
	"Ah, the epitome of incompetence has graced us with their presence. Your one-click victory serves as a reminder of your astounding insignificance.",
	"Well, slap my face and call me a fool! Who knew that a game designed to challenge the mind could be won by sheer happenstance?",
	"Behold, the great enigma of our time! The person who solved Minesweeper without breaking a sweat. I bow down to your boundless stupidity.",
	"Hark! The herald of mindlessness has arrived. You have achieved the impossible: making ignorance seem like an accomplishment.",
	"Ladies and gentlemen, witness the pinnacle of futility! Our champion has triumphed by doing absolutely nothing of substance.",
	"Oh, look, it's a walking testament to ineptitude! Your feat of clueless luck shall forever be remembered in the annals of pointlessness.",
	"Marvelous job, my dear friend! You've proven that even a blind squirrel can find a nut once in a blue moon.",
	"How breathtakingly ordinary! You've managed to turn a game of skill into a farcical display of mindless fortune. Bravo, I suppose.",
	"Congratulations on discovering the hidden secret to Minesweeper: blind guesswork. Your triumph shall be celebrated in the halls of absurdity.",
	"The stars must have aligned in your favor, or perhaps the universe just took pity on your incompetence. Either way, your victory is utterly inconsequential.",
	"I must commend your expertise in the realm of sheer dumb luck. It takes a special kind of imbecile to succeed without even trying.",
	"Well, if it isn't our resident master of mindlessness! Your accomplishment serves as a shining example of how to achieve nothing with great ease.",
	"Oh, the brilliance of your apathetic victory! You've managed to reduce a test of wit to a mere exercise in mindless chance. Bravo, simpleton.",
}

var SarcasticLostMessages = []string{
	"Oops, you lost. Maybe you should consider a career in Minesweeper demolition.",
	"Defeat snatched away from the jaws of victory. You have a real talent for failure.",
	"Well, that was a catastrophic loss. I'm sure you'll find solace in knowing that nobody expected anything more from you.",
	"Another loss? I'm starting to think Minesweeper is your personal kryptonite.",
	"Lost again? Maybe you should stick to games that don't require basic problem-solving skills.",
	"Congratulations! You managed to lose at Minesweeper. It's a remarkable accomplishment in its own right.",
	"Bravo! You failed spectacularly at Minesweeper. Your talent for disaster is truly awe-inspiring.",
	"Ah, another soul crushed by the mercilessness of Minesweeper. How delightful.",
	"Your loss at Minesweeper is a shining example of your innate ability to fall short. Keep up the good work!",
}

var SarcasticTimeOverMessages = []string{
	"That's enough time to watch a movie or regret every decision that led you to this point.",
	"I can't decide if your dedication is impressive or utterly baffling. Let's go with the latter.",
	"Congratulations on your remarkable ability to squander precious moments of your existence.",
	"Keep up the excellent work. Your commitment to unproductive pursuits is truly unparalleled!",
	" Your ability to waste substantial chunks of your life on insignificant endeavors is truly something to behold. Bravo!",
	"You still haven't finished? Really? It's official, you've surpassed all expectations in the art of inefficiency. Maybe consider a career in slowness.",
	"Did you fall asleep while playing Minesweeper? Your sloth-like pace suggests you did. Don't worry; your life is just as exciting.",
	"You've been playing Minesweeper for so long that even snails are mocking your speed. Keep digging deeper into your pit of incompetence.",
	"You reached the time limit. Your Minesweeper skills are truly out of this world... and not in a good way.",
	"Time's up! I hope you enjoyed your extended vacation in the world of Minesweeper. Now it's time to face reality.",
	"Wow, you really pushed the boundaries of time with your Minesweeper performance. It's a skill, really.",
	"Congratulations on your exceptional ability to waste time on Minesweeper. The world truly needed another master of procrastination.",
	"Your leisurely pace at Minesweeper is a testament to your commitment to inefficiency. Well done!",
}

var SarcasticGiveUpMessages = []string{
	// This message is SUPER toxic but I love it so I'm keeping it.
	"Oh, look at you, throwing in the towel like a weakling. Did you realize you're as worthless as those unmarked mines you couldn't handle? How pathetic. Maybe if your dad hadn't abandoned you, you would've learned some perseverance. But alas, here you are, a failure at a simple game. It's no wonder nobody wants to be around you. Go ahead, give up on Minesweeper, just like your dad did on you. You're a disappointment in more ways than one.",
	"Oh, look at you, giving up on Minesweeper like a quivering, spineless jellyfish. You were never cut out for this game, or anything remotely challenging, for that matter.",
	"Ah, the sound of your defeat is like music to my ears. It's as if the universe itself rejoices at your inevitable failure. Well done.",
	"Quitting Minesweeper suits you perfectly, just like being a perpetual disappointment. I hope you enjoy your life of mediocrity.",
	"Your decision to quit Minesweeper is an accurate reflection of your overall insignificance. You're a tiny blip on the radar of life, easily erased and forgotten.",
	"Did you really think you could conquer Minesweeper? How naive of you. You're nothing more than a feeble-minded insect scurrying away from a challenge.",
	"Pathetic. Giving up on Minesweeper shows that you lack the mental fortitude to face even the simplest of tasks. Your weakness knows no bounds.",
	"Oh, look at you, quitting Minesweeper. It's as if your incompetence is a superpower, capable of sabotaging your every endeavor. Bravo, imbecile.",
	"Your inability to solve Minesweeper reveals your true nature: a hapless fool stumbling through life, leaving a trail of disappointment wherever you go.",
	"I must admit, your knack for quitting is truly impressive. It's almost as if you were born with a talent for surrendering. Well done, you quitter extraordinaire.",
	"Did you honestly think you could outsmart Minesweeper? It's clear that your intellectual capacity is on par with a brainless amoeba. Carry on, simpleton.",
	"Ah, another weakling succumbs to the challenges of Minesweeper. Your surrender is like a beacon of ineptitude, guiding others toward your pitiful example.",
	"Quitting Minesweeper is just another chapter in the grand book of your failures. I wonder, how many more pages will be filled with your ineptitude.",
	"Look at you, retreating from Minesweeper with your tail between your legs. Your defeat is a spectacle that brings me immense satisfaction. Carry on, loser.",
	"Congratulations on your brilliant decision to quit Minesweeper. It's an eloquent reminder of your perpetual underachievement and lack of ambition.",
	"The fact that you quit Minesweeper doesn't surprise me in the least. After all, why bother striving for success when you can revel in your comfortable failure.",
	"Quitting Minesweeper was a wise choice for someone of your caliber. It's clear that the world would be better off without your feeble attempts at accomplishment.",
	"Oh, the sweet taste of your failure. Quitting Minesweeper is just another example of how easily you crumble under even the mildest pressure. How amusing.",
	"Your decision to quit Minesweeper speaks volumes about your character. It reveals a distinct lack of perseverance and an overwhelming presence of incompetence.",
	"Ah, the dance of defeat. You quit Minesweeper, showcasing your remarkable talent for giving up when faced with the simplest of challenges. Bravo, imbecile.",
	"Quitting Minesweeper is a testament to your lack of resilience and determination. It's no wonder you're trapped in a cycle of perpetual disappointment and regret.",
	"Your choice to quit Minesweeper is the cherry on top of the towering mountain of your failures. I hope you find solace in your consistent mediocrity.",
}

func getRandomMessage(messages []string) string {
	index, _ := rand.Int(rand.Reader, big.NewInt(4))
	return messages[index.Int64()]
}

var timeOrder = []int64{int64(0), int64(5), int64(10), int64(15), int64(20), int64(30), int64(40), int64(50), int64(60), int64(70), int64(90), int64(60 * 2), int64(60 * 5), int64(60 * 10), int64(60 * 15), int64(60 * 20), int64(60 * 30), int64(60 * 40), int64(60 * 50)}

func getMessages(secondsSpent int64) []string {
	for index := len(timeOrder) - 1; index >= 0; index-- {
		minSeconds := timeOrder[index]
		if secondsSpent >= minSeconds {
			return SarcasticTimes[minSeconds]
		}
	}
	return []string{}
}
