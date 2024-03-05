package types

type SignatureAlgorithm string

const (
	SignatureAlgorithmRSAPSSSHA256      SignatureAlgorithm = "RSASSA_PSS_SHA_256"
	SignatureAlgorithmRSAPSSSHA384      SignatureAlgorithm = "RSASSA_PSS_SHA_384"
	SignatureAlgorithmRSAPSSSHA512      SignatureAlgorithm = "RSASSA_PSS_SHA_512"
	SignatureAlgorithmRSAPKCS1V15SHA256 SignatureAlgorithm = "RSASSA_PKCS1_V1_5_SHA_256"
	SignatureAlgorithmRSAPKCS1V15SHA384 SignatureAlgorithm = "RSASSA_PKCS1_V1_5_SHA_384"
	SignatureAlgorithmRSAPKCS1V15SHA512 SignatureAlgorithm = "RSASSA_PKCS1_V1_5_SHA_512"

	SignatureAlgorithmECDSASHA256 SignatureAlgorithm = "ECDSA_SHA_256"
	SignatureAlgorithmECDSASHA384 SignatureAlgorithm = "ECDSA_SHA_384"

	SignatureAlgorithmEd25519 SignatureAlgorithm = "Ed25519"
)
