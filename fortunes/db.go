package fortunes

// Guidelines for adding new fortunes:
// - Fortunes should be relevant to the Bitcoin community, and should contain
//   a nugget of wisdom or trivia about Bitcoin that satstack users will find
//   interesting.
//
// - Do not break lines, unless you really want an explicit line break. Please
//   see the existing examples for inspiration.
//
// - When mentioning the date, just including the year is enough.
//
// - When quoting someone, the text should start with a "-- " prefixed with a
//   tab character.
var bitcoinFortunes = []string{
	`It's very attractive to the libertarian viewpoint if we can explain it properly. I'm better with code than with words though.

	-- Satoshi Nakamoto, 2008`,

	`SHA-256 is very strong. It's not like the incremental step from MD5 to SHA1. It can last several decades unless there's some massive breakthrough attack.

	-- Satoshi Nakamoto, 2010`,

	`We can win a major battle in the arms race and gain a new territory of freedom for several years.

	-- Satoshi Nakamoto, 2008`,

	`As an additional firewall, a new key pair should be used for each transaction to keep them from being linked to a common owner.

	-- Satoshi Nakamoto, 2008`,

	`Banks must be trusted to hold our money and transfer it electronically, but they lend it out in waves of credit bubbles with barely a fraction in reserve. We have to trust them with our privacy, trust them not to let identity thieves drain our accounts. Their massive overhead costs make micropayments impossible.

	-- Satoshi Nakamoto, 2009`,

	`I've been working on a new electronic cash system that's fully peer-to-peer, with no trusted third party.

	-- Satoshi Nakamoto, 2008`,

	`I believe I've worked through all those little details over the last year and a half while coding it, and there were a lot of them.

	-- Satoshi Nakamoto, 2008`,

	`NYTimes 09/Apr/2020 With $2.3T Injection, Fed's Plan Far Exceeds 2008 Rescue

	Block 629999 coinbase
	The last 12.5 BTC block`,

	`I'm sure that in 20 years there will either be very large transaction volume or no volume.

	-- Satoshi Nakamoto, 2010`,

	`Bitcoins have no dividend or potential future dividend, therefore not like a stock. More like a collectible or commodity.

	-- Satoshi Nakamoto, 2010`,

	`Running bitcoin

	-- Hal Finney, 2009`,

	`The Times 03/Jan/2009 Chancellor on brink of second bailout for banks

	Block 0 coinbase (genesis)`,

	`I'll pay 10,000 bitcoins for a couple of pizzas.. like maybe 2 large ones so I have some left over for the next day.

	-- laszlo (on bitcointalk), 2010`,

	`I AM HODLING

	-- @GameKyuubi (on bitcointalk), 2013`,

	`To me it looks like an impressive job, although I'd wish for more comments. Now I've mostly studied the init, main, script and a bit of net modules. This is some powerful machinery.

	-- Hal Finney, 2010`,

	`[I've been working on this design] since 2007. At some point I became convinced there was a way to do this without any trust required at all and couldn't resist to keep thinking about it. Much more of the work was designing than coding.

	-- Satoshi Nakamoto, 2010`,

	`Being open source means anyone can independently review the code. If it was closed source, nobody could verify the security. I think it's essential for a program of this nature to be open source.

	-- Satoshi Nakamoto, 2009`,

	`We define an electronic coin as a chain of digital signatures. Each owner transfers the coin to the next by digitally signing a hash of the previous transaction and the public key of the next owner and adding these to the end of the coin. A payee can verify the signatures to verify the chain of ownership.

	-- Satoshi Nakamoto, 2010`,

	`For greater privacy, it's best to use bitcoin addresses only once.

	-- Satoshi Nakamoto, 2009`,

	`When Satoshi announced the first release of the software, I grabbed it right away. I think I was the first person besides Satoshi to run bitcoin.

	-- Hal Finney, 2013`,

	`I thought I was dealing with a young man of Japanese ancestry who was very smart and sincere. I've had the good fortune to know many brilliant people over the course of my life, so I recognize the signs.

	-- Hal Finney, 2013`,

	`Those were the days when difficulty was 1, and you could find blocks with a CPU, not even a GPU. I mined several blocks over the next days. But I turned it off because it made my computer run hot, and the fan noise bothered me.

	-- Hal Finney, 2013`,

	`Writing a description for this thing [Bitcoin] for general audiences is bloody hard.  There's nothing to relate it to.

	-- Satoshi Nakamoto, 2010`,

	`Governments are good at cutting off the heads of a centrally controlled networks like Napster, but pure P2P networks like Gnutella and Tor seem to be holding their own.

	-- Satoshi Nakamoto, 2009`,

	`The root problem with conventional currency is all the trust that's required to make it work. The central bank must be trusted not to debase the currency, but the history of fiat currencies is full of breaches of that trust.

	-- Satoshi Nakamoto, 2009`,

	`I've developed a new open source P2P e-cash system called Bitcoin. It's completely decentralized, with no central server or trusted parties, because everything is based on crypto proof instead of trust.

	-- Satoshi Nakamoto, 2009`,

	`Difficulty just increased by 4 times, so now your cost is US$0.02/BTC.

	-- Satoshi Nakamoto, 2010`,

	`In a few decades when the reward gets too small, the transaction fee will become the main compensation for nodes.

	-- Satoshi Nakamoto, 2010`,

	`Lost coins only make everyone else's coins worth slightly more. Think of it as a donation to everyone.

	-- Satoshi Nakamoto, 2010`,

	`Well this is an exceptionally cute idea, but there is absolutely no way that anyone is going to have any faith in this currency.

	-- Joe Doliner (first public doubter of Bitcoin, on HN), 2009`,

	`It's a pyramid scheme. You only make money based on people who enter after you. It has no real utility in the world.

	-- Tendayi Kapfidze (economist), 2020`,

	`I won't be talking about Bitcoin in 10 years, I can assure you that. [...] I would bet in even 5 or 6 years I'm no longer talking about Bitcoin as Treasury Secretary.

	-- Steven Mnuchin (US Secretary of the Treasury), 2019`,

	`It’s a terrible store of value. It could be replicated over and over.

	-- Jamie Dimon (President & CEO, JPMorgan Chase), 2014`,

	`Stay away from it. It’s a mirage basically.

	-- Warren Buffett (Chairman & CEO, Berkshire Hathaway), 2014`,

	`I am not Dorian Nakamoto.

	-- Satoshi Nakamoto, 2014`,

	`A purely peer-to-peer version of electronic cash would allow online payments to be sent directly from one party to another without going through a financial institution.

	-- Satoshi Nakamoto, 2008`,

	`I would be surprised if 10 years from now we're not using electronic currency in some way, now that we know a way to do it that won't inevitably get dumbed down when the trusted third party gets cold feet.

	-- Satoshi Nakamoto, 2009`,

	`With e-currency based on cryptographic proof, without the need to trust a third party middleman, money can be secure and transactions effortless.

	-- Satoshi Nakamoto, 2009`,

	`A lot of people automatically dismiss e-currency as a lost cause because of all the companies that failed since the 1990's. I hope it's obvious it was only the centrally controlled nature of those systems that doomed them. I think this is the first time we're trying a decentralized, non-trust-based system.

	-- Satoshi Nakamoto, 2009`,

	`The utility of the exchanges made possible by Bitcoin will far exceed the cost of electricity used. Therefore, not having Bitcoin would be the net waste.

	-- Satoshi Nakamoto, 2010`,

	`Instead of the supply changing to keep the value the same, the supply is predetermined and the value changes. As the number of users grows, the value per coin increases. It has the potential for a positive feedback loop; as users increase, the value goes up, which could attract more users to take advantage of the increasing value.

	-- Satoshi Nakamoto, 2009`,

	`You could say coins are issued by the majority. They are issued in a limited, predetermined amount.

	-- Satoshi Nakamoto, 2009`,

	`The price of any commodity tends to gravitate toward the production cost. If the price is below cost, then production slows down. If the price is above cost, profit can be made by generating and selling more. At the same time, the increased production would increase the difficulty, pushing the cost of generating towards the price.

	-- Satoshi Nakamoto, 2010`,

	`The project needs to grow gradually so the software can be strengthened along the way. I make this appeal to WikiLeaks not to try to use Bitcoin. Bitcoin is a small beta community in its infancy.

	-- Satoshi Nakamoto, 2010`,

	`The requirement is that the good guys collectively have more CPU proof-of-worker than any single attacker.

	-- Satoshi Nakamoto, 2008`,

	`The proof-of-work chain is the solution to the synchronisation problem, and to knowing what the globally shared view is without having to trust anyone.

	-- Satoshi Nakamoto, 2008`,

	`The credential that establishes someone as real is the ability to supply CPU proof-of-worker.

	-- Satoshi Nakamoto, 2008`,

	`Commerce on the Internet has come to rely almost exclusively on financial institutions serving as trusted third parties to process electronic payments. While the system works well enough for most transactions, it still suffers from the inherent weaknesses of the trust based model.

	-- Satoshi Nakamoto, 2008`,

	`Bitcoin is outdated technology - almost prehistoric by crypto standards.

	-- Mike Hearn, 2011`,

	`Despite knowing that bitcoin could fail all along, the now inescapable conclusion that it has failed still saddens me greatly.

	-- Mike Hearn, 2011`,
}
