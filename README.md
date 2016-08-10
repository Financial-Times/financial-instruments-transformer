Financial Instruments transformer
=================================

[![Circle CI](https://circleci.com/gh/Financial-Times/financial-instruments-transformer/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/financial-instruments-transformer/tree/master)

Transforms factset security data files into Financial Instruments

How to Build & Run the binary
-----------------------------

1. Build and test:

        go build
        go test

2. Run:

        export AWS_SECRET_ACCESS_KEY="***" \
            && export AWS_ACCESS_KEY_ID="***" \
            && export BUCKET_NAME="com.ft.coco-factset-data" \
            && ./financial-instruments-transformer

Notes
-----

As of today (10th of August) the transformer resolves multiple entities/securities pointing to the same FIGI by choosing the record which is non-expired (has no termination date)
If there are more records with no termination date, one randomly will be picked.  
