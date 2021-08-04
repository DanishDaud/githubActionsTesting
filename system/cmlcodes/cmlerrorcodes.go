package cmlcodes

// Error Codes for Contact List
const (
	ContactListErrorNoError                      = 500
	ContactListErrorNumberColumnInvalid          = 501
	ContactListErrorOpeningFileForDataExtraction = 502
	ContactListErrorContactGroupCreation         = 503
	ContactListErrorUnknown                      = 599

	CampaignErrorCGsNotInRedis                    = 700
	CampaignErrorCGsFetchDbFailed                 = 701
	CampaignErrorCGsNosPushToRedisFailed          = 702
	CampaignErrorPendingNosCountRedisFailed       = 703
	CampaignErrorStateFetchedFromRedisFailed      = 704
	CampaignErrorSoundFileInfoFetchFailed         = 705
	CampaignErrorVMSoundFileInfoFetchFailed       = 706
	CampaignErrorDNCSoundFileInfoFetchFailed      = 707
	CampaignErrorTransferSoundFileInfoFetchFailed = 708
	CampaignErrorAnySoundFileInfoFetchFailed      = 709
	CampaignErrorCampaignUserFetchFailed          = 710
	CampaignErrorCampaignParentUserFetchFailed    = 711
	NoError                                       = 200

	CampaignStop         = 999
	CampaignScheduleStop = 900
)
