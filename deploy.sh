#!/bin/bash

gcloud run deploy --region=asia-northeast1 --source=. --set-env-vars=PROJECT_ID=$PROJECT_ID --allow-unauthenticated my-app
