package cmlmessages

// Messages Codes System
const (
	CarrierError          = "Carrier Error"
	PleaseTryLater        = "Please Try Later"
	PleaseTryAgain        = "Please Try Again"
	PleaseEnterDigitAgain = "Please Enter 10 digit number"
	NumberNotFound        = "Number Not Found In Carrier"
	OperationFailed       = "Operation Failed"
	InvalidObjectId       = "Invalid Object Id"
	UnauthorizedForAction = "You are unauthorized to perform this action"
)

// Messages Codes for Users
const (
	CompulsoryFieldsMissing    = "Please provide all compulsory fields"
	RequestFormatIncorrect     = "Request Format is incorrect"
	UserRequestFormatIncorrect = "User Request Format is incorrect"
	UserRoleIncorrect          = "Requested role for user can not be assigned"
	SuperUserAlreadyExists     = "Super User is already present"
	UserWithEmailAlreadyExists = "User with this email address already exists"
	InvalidParentId            = "Parent Id is invalid"
	UserCreationFailed         = "Failed to Create User"
	DomainOwnerDoesNotExist    = "No owner exists of this domain"
	UserDoesNotExist           = "User does not exist"
	UserOrPasswordDoesNotExist = "User name or password is incorrect"
	UserGmailIdIsInvalid       = "Please provide valid gmail account id"
	UserOldPasswordMismatch    = "Please provide correct old password"
	UserAccountLocked          = "Your account is locked"
)

// Messages Codes for Users
const (
	SoundFileUploadFail                = "Sound file failed to upload"
	SoundFileReadError                 = "Failed to read uploaded sound file"
	SoundFileWriteError                = "Failed to write uploaded sound file"
	SoundFileFormatInCorrect           = "Only .mp3 and .wav files are supported"
	SoundFileNameUpdateFormatIncorrect = "Incorrect format"
	SoundFileUpdateNameMissing         = "Please provide name to update"
	SoundFileDoesNotExist              = "Sound file does not exist"
	SoundFileNoAccessToUpdate          = "Sorry! You don't have access rights to make this change"
	SoundFileDeleteFailForS3           = "Sound file not present on s3"
	SoundFileDeleteOperationFailed     = "Failed to delete sound file"
	SoundFileUnauthorizedForAction     = "You are unauthorized to perform this action"
	SoundFileAttachedWithCamapign      = "Sound File is attached to a campaign, you can not delete it"
	SoundFileInputFormatIncorrect      = "Please provide valid input"
)

// Messages Codes for ContactList
const (
	ContactListSelectFile            = "Please provide contact list"
	ContactListReadError             = "Failed to read uploaded file"
	ContactListWriteError            = "Failed to write uploaded file"
	ContactListFormatInCorrect       = "Only .csv, .xls and .xlsx files are supported"
	ContactListNumberColumnIncorrect = "Please mention correct number column"
	ContactListTextColumnIncorrect   = "Please mention correct text  column"
	ContactListProvideAllFields      = "Please provide all required fields"
	ContactListDoesNotExist          = "Contact List does not exist"
	TTSContactListDoesNotExist       = "TTS Contact List does not exist"
	ContactListUnauthorizedForAction = "You are unauthorized to perform this action"
	ContactListDeleteFailForS3       = "Contact list not present on s3"
	ContactListDeleteOperationFailed = "Contact list delete sound file"
	ContactListAttachedWithCamapign  = "Contact list is attached to a campaign, you can not delete it"
	ContactListInvalidContactNumber  = "Invalid contact number"
)

// Messages Codes for Campaign
const (
	CampaignRequestFormatIncorrect = "Campaign request format is incorrect"
	CampaignSaveFail               = "Failed to save campaign"
	CampaignDeleteFail             = "Campaign can not be deleted"
	CampaignDoesNotExist           = "Campaign does not exist"
	CampaignRunAlreadyPresent      = "Campaign run is already present"
	CampaignCantUpdateRunning      = "Running campaign can not be updated"
)
