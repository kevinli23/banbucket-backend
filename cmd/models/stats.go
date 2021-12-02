package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
)

type FirebaseTransaction struct {
	Hash      string  `json:"hash"`
	Address   string  `json:"address"`
	Amount    float64 `json:"amount"`
	Height    uint32  `json:"height"`
	Timestamp uint32  `json:"timestamp"`
}

type YellowGlassTransactionsResponse struct {
	Hash       string `json:"hash"`
	Address    string `json:"address"`
	Type       string `json:"type"`
	BalanceRaw string `json:"balanceRaw"`
	Height     uint32 `json:"height"`
	Timestamp  uint32 `json:"timestamp"`
}

type FireBaseTransactionMetadataResponse struct {
	Offset      uint32 `json:"offset"`
	TotalClaims uint32 `json:"totalclaims"`
}

type BanBucketStats struct {
	TotalClaims        uint32            `json:"total_claims,omitempty"`
	ClaimsToday        uint32            `json:"today_claims,omitempty"`
	ClaimsYesterday    uint32            `json:"yesterday_claims,omitempty"`
	AverageDailyClaims uint32            `json:"average_daily_claims,omitempty"`
	UniqueClaims       uint32            `json:"unique_claims,omitempty"`
	TotalSent          float32           `json:"total_sent,omitempty"`
	DailyClaims        map[string]uint32 `json:"daily_claims,omitempty"`
	LastUpdate         uint64            `json:"last_updated,omitempty"`
}

type FirestoreHandler struct {
	Client      *firestore.Client
	CachedStats BanBucketStats `json:"stats"`
}

func (f *FirestoreHandler) New() {
	ctx := context.Background()

	jsonCredentials, _ := json.Marshal(map[string]string{
		"type":                        "service_account",
		"project_id":                  os.Getenv("project_id"),
		"private_key_id":              os.Getenv("private_key_id"),
		"private_key":                 os.Getenv("private_key"),
		"client_email":                os.Getenv("client_email"),
		"client_id":                   os.Getenv("client_id"),
		"auth_uri":                    os.Getenv("auth_uri"),
		"token_uri":                   os.Getenv("token_uri"),
		"auth_provider_x509_cert_url": os.Getenv("auth_provider_x509_cert_url"),
		"client_x509_cert_url":        os.Getenv("client_x509_cert_url"),
	})

	sa := option.WithCredentialsJSON(jsonCredentials)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	f.Client = client
	f.CachedStats = BanBucketStats{}
}

// Returns the offset and totalclaims
func (f *FirestoreHandler) GetMetadata() (uint32, uint32, error) {
	ctx := context.Background()

	doc, _ := f.Client.Collection("metadata").Doc("info").Get(ctx)

	var res FireBaseTransactionMetadataResponse

	jsonBytes, _ := json.Marshal(doc.Data())

	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		return 0, 0, err
	}

	return res.Offset, res.TotalClaims, nil
}

func (f *FirestoreHandler) SetMetadata(offset uint32, totalclaims uint32) error {
	ctx := context.Background()
	doc := f.Client.Collection("metadata").Doc("info")

	doc.Set(ctx, map[string]interface{}{
		"offset":      offset,
		"totalclaims": totalclaims,
	})

	return nil
}

func (f *FirestoreHandler) GetStats() BanBucketStats {
	return f.CachedStats
}

func (f *FirestoreHandler) GenerateStats(ctx context.Context, collection *mongo.Collection) error {
	_, total, err := f.GetMetadata()
	if err != nil {
		return errors.Wrap(err, "Failed to get metadata")
	}

	uniqueClaims, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return errors.Wrap(err, "Failed to count mongo documents")
	}

	// Grab today's document
	year, month, day := time.Now().Date()
	today := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)

	year, month, day = time.Now().AddDate(0, 0, -1).Date()
	yesterday := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)

	allDocumentsRef := f.Client.Collection("transactions").DocumentRefs(ctx)

	documentCounts := map[string]uint32{}

	allDocuments, err := allDocumentsRef.GetAll()
	if err != nil {
		return err
	}

	for _, docRefs := range allDocuments {
		snapshot, err := docRefs.Get(ctx)
		if err != nil {
			return err
		}

		documentCounts[docRefs.ID] = uint32(snapshot.Data()["count"].(int64))
	}

	// TODO: Daily Claims
	f.CachedStats.TotalClaims = total
	f.CachedStats.UniqueClaims = uint32(uniqueClaims)
	f.CachedStats.ClaimsToday = documentCounts[today]
	f.CachedStats.ClaimsYesterday = documentCounts[yesterday]
	f.CachedStats.DailyClaims = documentCounts
	f.CachedStats.LastUpdate = uint64(time.Now().Unix())

	return nil
}
