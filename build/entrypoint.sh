#!/bin/bash
#entrypoint script for the docker file to run the application :D




# Step 1: Download the latest release
echo "Step 1: Downloading the latest release"
mkdir -p /app
curl  https://github.com/coreweave/gutenberg-epub/releases/latest/download/gutenberg-epub -O /app/gutenberg-epub-converter
chmod +x /app/gutenberg-epub-converter

# Step 2: Download files from S3 using s3cmd
echo "Step 2: Downloading files from S3"
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey get s3://gutenberg/pg-calibre-library

# Step 3: Process the files (replace this with your actual processing logic)
echo "Step 3: Processing files"
./gutenberg-epub-converter -inputDir ./pg-calibre-library -outputDir ./gutenberg-by-author -writeHeader=true -writeMetadata=false -cleanOutput=true -seperateFolders=false -stopEarly=0 -skipCopyRight=true -gutenbergCleaning=true -createSubsets=author


# Step 4: Upload results to S3 using s3cmd
echo "Step 4: Uploading results to S3"
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey --no-check-hostname --no-check-certificate put ./gutenberg-by-author s3://gutenberg --recursive --multipart-chunk-size-mb=50 -H --progress --stats 

# Step 5: Return a "done" message
echo "Done: Files processed and results uploaded to S3"
