package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	sal "github.com/salrashid123/oauth2/google"
	"google.golang.org/api/option"
)

var (
	gcpBucket               = flag.String("gcpBucket", "mineral-minutia-820-cab1", "GCS Bucket to access")
	gcpObjectName           = flag.String("gcpObjectName", "foo.txt", "GCS object to access")
	gcpResource             = flag.String("gcpResource", "//iam.googleapis.com/projects/1071284184436/locations/global/workloadIdentityPools/oidc-pool-1/providers/oidc-provider-1", "the GCP resource to map")
	gcpTargetServiceAccount = flag.String("gcpTargetServiceAccount", "oidc-federated@mineral-minutia-820.iam.gserviceaccount.com", "the ServiceAccount to impersonate")

	sourceToken = flag.String("sourceToken", "", "Source OIDC token to echange")

	scope = flag.String("scope", "https://www.googleapis.com/auth/cloud-platform", "Scope of the target token")

	useIAMToken = flag.Bool("useIAMToken", true, "Use IAMCredentials Token exchange")
)

func main() {
	flag.Parse()

	if *sourceToken == "" {
		log.Fatalf("sourceToken cannot benull")
	}

	oTokenSource, err := sal.OIDCFederatedTokenSource(
		&sal.OIDCFederatedTokenConfig{
			SourceToken:          *sourceToken,
			Scope:                *scope,
			TargetResource:       *gcpResource,
			TargetServiceAccount: *gcpTargetServiceAccount,
			UseIAMToken:          *useIAMToken,
		},
	)

	tok, err := oTokenSource.Token()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("OIDC Derived GCP access_token: %s\n", tok.AccessToken)

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx, option.WithTokenSource(oTokenSource))
	if err != nil {
		log.Fatalf("Could not create storage Client: %v", err)
	}

	bkt := storageClient.Bucket(*gcpBucket)
	obj := bkt.Object(*gcpObjectName)
	r, err := obj.NewReader(ctx)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	if _, err := io.Copy(os.Stdout, r); err != nil {
		panic(err)
	}

}
