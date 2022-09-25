package main

//noinspection GoUnusedConst
const (
	scbc             = "624419155625639937"
	nevana           = "527668318451007528"
	prettyembedcolor = 0x02346c
	errorColor       = 0xff2222

	info          = "624423595938152448"
	rules         = "625447581392044032"
	announcements = "624423646311481384"
	partners      = "625498052484005894"
	events        = "625497751051829269"
	adminChat     = "625835839364268072"
	scbcPosts     = "929476853218033754"
	modLog        = oldRoles

	general = "686929018968211459"
	memes   = "687437401919848504"
	media   = "692442671683469312"

	botCommands                         = "692442563491397642"
	roles                               = "933914370952532009"
	oldRoles                            = "692442579186221167"
	selfPromotion                       = "692442610765135892"
	stupidDiscordCommunityAnnouncements = "779451633456250931"

	woodwind   = "627257810492588032"
	colorguard = "627257871398076426"
	percussion = "627257828704256021"
	brass      = "627257674911711260"

	wandoKids = "686747343265071132"
)

var (
	Admin         = Role{"624423197592256512", "Admin"}
	Mod           = Role{"625659968120422400", "Mod"}
	SCBandChatorg = Role{"641267217262051379", "SCBandChat.org"}
)

type Role struct {
	// Discord id
	ID   string
	Name string
}

func RolesToIDs(roles []Role) []string {
	var ids []string
	for _, role := range roles {
		ids = append(ids, role.ID)
	}
	return ids
}

var muteRoles = map[string]string{
	"":            "653784746308010005",
	general:       "937499112344023050",
	media:         "937499263854854215",
	memes:         "937499313876123688",
	woodwind:      "937499351687770122",
	colorguard:    "937499395723776101",
	percussion:    "937499444595814420",
	brass:         "937499491949510696",
	selfPromotion: "953009007923167303",
}

// ActivityTypeGame      ActivityType = 0
// ActivityTypeStreaming ActivityType = 1
// ActivityTypeListening ActivityType = 2
// ActivityTypeWatching  ActivityType = 3
// ActivityTypeCustom    ActivityType = 4
// ActivityTypeCompeting ActivityType = 5

var activityTypes = map[int]string{
	0: "playing",
	//1: "streaming",
	2: "listening",
	3: "watching",
	//4: "custom",
	5: "competing",
}

//You only want to use types 0, 2, 3, and 5. The others aren't allowed to be used by bots. See above what each type refers to

var splashText = map[int]string{
	3: /*Watching*/ "a bunch of nerds",
}

var schools1 = []string{
	"626921356813926430", //Blue Ridge
	"624423835005091880", //Blythewood
	"632962238847778826", //Boiling Springs
	"844767462929989643", //Carolina Forest
	"904911749717827594", //Catawba Ridge
	"632967288785731595", //Chapman
	"976266915758633050", //Chapin
	"652665462898819072", //Chesnee
	"657275633935712258", //Clover
	"625518206408327168", //DW Daniels
	"666486159592914944", //Dorman
	"653045716813479938", //Dreher
	"624424875297210379", //Dutch Fork
	"653029291260510209", //Easly
	"624458714795212801", //Edisto
	"1022639629675008110", //Fort Dorchester
	"625489235624984576", //Fort Mill
	"844766300999647252", //Greenville
	"632963567557148673", //Hartsville
	"903459757959053312", //Irmo
	"627249596317302785", //James F Byrnes
	"653028895456624640", //JL Mann
	"625870330388545555", //Laurens
	"625647039002312715", //Lexington
	"1023396806081658940"}//Liberty
	

var schools2 = []string{
	"895120511003132016", //Mauldin
	"636900835875487774", //Nation Ford
	"625460205756350485", //North Agusta
	"903808150304620605", //Pelion
	"901890027288670228", //Pendleton
	"625840098210086923", //RiverBluff
	"636908685813350401", //Riverside
	"637810429418274826", //Spartanburg
	"637651502814461953", //Spring Valley
	"755601893652168854", //Stratford
	"625511435400642563", //Summerville
	"637810048063766569", //Wade hampton
	"626535659288395806", //Wagener-Salley
	"640166574694596608", //White Knoll
	"983170639865184296", //Wren
	"624424700680077313"} //York

var classes = []string{
	"631653089123762216", //Graduate
	"625861140169228288", //senior
	"625861196548931594", //Junior
	"625861225066135564", //Sophomore
	"625861249531510806", //Freshman
	"625861277641867275"} //8th Grade

var schoolSize = []string{
	"631652081161469978", //6a
	"625501413128273933", //5a
	"625501441897005066", //4a
	"625501476487168000", //3a
	"625501500239642637", //2a
	"625501525359460352"} //1a

var instruments1 = []string{
	"627258126256832513", //Colorguard
	"933934314146631790", //Piccolo
	"625476668320120833", //Flute
	"625476617631694878", //Clarinet
	"910021771879661619", //Bass Clarinet
	"934299687979221062", //Soprano Saxophone
	"627258738415501323", //Alto Saxophone
	"933934882466430986", //Tenor Saxophone
	"635276668684206100", //Bari Saxophone
	"625857942989701139", //Trumpet
	"625870956715442186", //Mellophone
	"626555254116188160", //Trombone
	"631655976822636575", //Bass Trombone
	"625872778909974558", //Baritone
	"625488729137741824", //Tuba
	"771818922705420288", //Strings
	"894432396986970142", //Bass Guitar
	"933934671115472906", //Glockenspiel
	"627259133007233054", //Xylophone
	"625476578926657536", //Vibrophone
	"625476488807972864", //Marimba
	"627259104615727134", //Bells
	"626588091246968832", //Rack
	"626202774165651457", //Synth
	"625505103822192661", //Snare Drum
}

var instruments2 = []string{
	"625505150559059988", //Tenor Drums
	"625483605585559552", //Bass Drums
	"661966954340417546"} //Cymbals

var leadership = []string{
	"625498161854808085", //Drum major
	"627318405250154507", //Section Leader
	"627901888704020491"} //Band Staff member

func (r Role) Mention() string {
	return "<@&" + r.ID + ">"
}
