#!/bin/bash

gcloud run deploy --region=asia-northeast1 --source=. --set-env-vars=PROJECT=$PROJECT --allow-unauthenticated my-app-golang
