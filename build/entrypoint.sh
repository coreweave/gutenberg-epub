#!/bin/bash
#entrypoint script for the docker file to run the application :D




# Step 1: Download the latest release
echo "Step 1: Downloading the latest release"
mkdir -p /app
echo $gh_token | gh auth login --with-token
echo "Authenticated. Downloading latest release"
gh release download --repo coreweave/gutenberg-epub -p '*'
chmod +x ./gutenberg-epub-converter
[ ! -f ./gutenberg-epub-converter ] && echo "Binary not found! Exiting"  && exit 1|| echo "Binary found!"

# Step 2: Download files from S3 using s3cmd
echo "Step 2: Downloading files from S3"
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey  --no-check-hostname --no-check-certificate la
mkdir ./pg-calibre-library && cd ./pg-calibre-library #s3cmd removes 1 layer of folders when downloading
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey  --no-check-hostname --no-check-certificate --recursive --quiet get s3://gutenberg/pg-calibre-library/
cd ..

# Step 3: Process the files (replace this with your actual processing logic)
echo "Step 3: Processing files"
ls -la
[ ! -d ./pg-calibre-library ] && echo "Library not found! Exiting"  && exit 1|| echo "Library found!"
./gutenberg-epub-converter -inputDir ./pg-calibre-library -outputDir ./gutenberg-by-author -writeHeader=true -writeMetadata=false -cleanOutput=true -seperateFolders=false -stopEarly=0 -skipCopyRight=true -gutenbergCleaning=true -createSubsets=author


# Step 4: Upload results to S3 using s3cmd
echo "Step 4: Uploading results to S3"
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey --no-check-hostname --no-check-certificate put ./gutenberg-by-author s3://gutenberg --recursive --multipart-chunk-size-mb=50 -H --progress --stats --quiet
s3cmd --access_key $s3accesskey --host $s3base --host-bucket $s3bucket --secret_key $s3secretkey  --no-check-hostname --no-check-certificate la

# Step 5: Return a "done" message
echo "Done: Files processed and results uploaded to S3"
