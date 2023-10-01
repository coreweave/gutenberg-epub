#!/bin/bash
#entrypoint script for the docker file to run the application :D




# Step 1: Download the latest release
echo "Step 1: Downloading the latest release"
mkdir -p /app
echo $gh_token | gh auth login --with-token
./bin/gh release download --repo coreweave/gutenberg-epub -p '*'
chmod +x ./gutenberg-epub-converter
[ ! -f /tmp/foo.txt ] && echo "File not found!"  && exit 1|| echo "File found!"

# Step 2: Download files from S3 using s3cmd
echo "Step 2: Downloading files from S3"
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey  --no-check-hostname --no-check-certificate la
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey  --no-check-hostname --no-check-certificate --recursive get s3://gutenberg/pg-calibre-library/

# Step 3: Process the files (replace this with your actual processing logic)
echo "Step 3: Processing files"
./gutenberg-epub-converter -inputDir ./pg-calibre-library -outputDir ./gutenberg-by-author -writeHeader=true -writeMetadata=false -cleanOutput=true -seperateFolders=false -stopEarly=0 -skipCopyRight=true -gutenbergCleaning=true -createSubsets=author


# Step 4: Upload results to S3 using s3cmd
echo "Step 4: Uploading results to S3"
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey --no-check-hostname --no-check-certificate put ./gutenberg-by-author s3://gutenberg --recursive --multipart-chunk-size-mb=50 -H --progress --stats 

# Step 5: Return a "done" message
echo "Done: Files processed and results uploaded to S3"
