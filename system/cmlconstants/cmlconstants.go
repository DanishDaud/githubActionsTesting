package cmlconstants

// ip to serve api on
//const ServeIP = "localhost:3000" // for local
const ServeIP = "0.0.0.0:3000" // for public webserver

// s3 bucket name
// all the s3 paths would be saved relative to bucket name in project
//const S3BucketName = "resources.callmylist.com"
//const S3BucketFullPath = "http://resources.callmylist.com"
//const VoipServerApiV1 = "http://api1.mycallblast.com:3000/v1/"
//const VoipServerApiLocalV1 = "http://localhost:3000/v1/"
//const CdrService = "http://api1.mycallblast.com:8085/api/"

// path to save files temporary to
// DO REMEMBER TO DELETE FILES SAVED HERE AFTER USER
const TempDestinationPath = "./Temp/"

const (
	ConfigContactListBatchSize = 10000
	TCPABatchSize              = 3000
)

//// campaign types
//const (
//	CampaignTypeVoiceOnly                              = 1
//	CampaignTypeLiveAnswerAndAnswerMachineNoTransfer   = 2
//	CampaignTypeLiveAnswerAndAnswerMachineWithTransfer = 3
//	CampaignTypeLiveAnswerOnlyNoTransfer               = 4
//	CampaignTypeLiveAnswerOnlyWithTransfer             = 5
//	CampaignTypeTextBlast                              = 6
//)

// campaign Statuses
const (
	CampaignStatusNew           = 1
	CampaignStatusRunning       = 2
	CampaignStatusStopped       = 3
	CampaignStatusFinished      = 4
	CampaignStatusError         = 5
	CampaignStatusScheduledStop = 6
	CampaignStatusPaused        = 7
)

// Contract Type
const (
	ContractTypePayAsYouGo   = 1
	ContractTypeMonthly      = 2
	ContractTypeQuarterly    = 3
	ContractTypeSemiAnnually = 4
	ContractTypeAnnually     = 5
)

// Billing Type
const (
	BillingTypePerContact = 1
	BillingTypePer6Secs   = 2
	BillingTypePer30Secs  = 3
	BillingTypePerMinute  = 4
)

// campaign default limits
const CampaignDefaultCCLimit = 50
const CampaignDefaultMaxTransferLimit = 50

// payment method
const PaymentMethodWePay = 1
const PaymentMethodStripe = 2
