Financial Instruments Transformer
=================================

[![Circle CI](https://circleci.com/gh/Financial-Times/financial-instruments-transformer/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/financial-instruments-transformer/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/financial-instruments-transformer)](https://goreportcard.com/report/github.com/Financial-Times/financial-instruments-transformer)

Transforms factset security data files into Financial Instruments.

This API is not publicly accessible.

How to Build & Run the binary
-----------------------------

1. Build and test:

        go build
        go test

2. Run:

        export AWS_SECRET_ACCESS_KEY="***" \
            && export AWS_ACCESS_KEY_ID="***" \
            && export BUCKET_NAME="com.ft.coco-factset-data" \
            && export BASE_URL="http://myhost/transformers/financial-instruments/" \
            && ./financial-instruments-transformer

Endpoints
----------

### GET
1. /transformers/financialinstruments/{uuid}: reads the financial instrument with the given uuid. A not found financial instrument will result in a 404 status code response.

`curl -H "X-Request-Id: 123" localhost:8080/transformers/financial-instruments/11f5ccf1-e6bf-3ec6-abaf-6380009a6c4b`

Successful response:
    * status code: 200
    * body: `{"uuid":"11f5ccf1-e6bf-3ec6-abaf-6380009a6c4b","prefLabel":"SAGA COMMUNICATIONS INC  CL A","alternativeIdentifiers":{"uuids":["11f5ccf1-e6bf-3ec6-abaf-6380009a6c4b"],"factsetIdentifier":"DCZBY8-S-US","figiCode":"BBG000F9R281"},"issuedBy":"3aa12e48-8835-30d2-9ed9-606447ebd36a"}`
    
2. /transformers/financial-instruments/__ids: reads the IDs of the financial instruments.

Successful response:
    * status code: 200
    * body: `{"id":"0c6842aa-e858-3053-b034-687e6db9578a"}\n{"id":"3bb726ff-7bf3-3303-8b09-caa226cdd208"}\n...`

3. /transformers/financial-instruments: reads the IDs of the financial instruments and returns them formatted as API URLs.

Successful response:
    * status code: 200
    * body: `[{"apiUrl":"http://<host>:8080/transformers/financial-instruments/bebcca96-a20e-3f38-9af9-88a4d008c3bb"},{"apiUrl":"http://<host>:8080/transformers/financial-instruments/e2bf1e03-7707-3ddd-b6b5-130064a02f63"},...\n]`

Admin endpoints
---------------
Health checks: http://localhost:8080/__health    
    
Notes
-----

As of today (10th of August) the transformer resolves multiple entities/securities pointing to the same FIGI by choosing the record which is non-expired (has no termination date)
If there are more records with no termination date, one randomly will be picked.  
