#!/bin/bash

########
# Help #
########
Help() {
        # Display Help
        echo "Convert yaml inside api/http and api/doc/swagger from openapi 3 to swagger 2."
        echo
        echo "Usage:"
        echo "  ./swaggerdoc.sh [options]"
        echo
        echo "options:"
        echo "h    Display Help"
        echo "b    Change base file (default: empty)"
        echo
}
files=""

while getopts ":hb:" flag;
do
        case "$flag" in
                h) Help
                   exit;;
                b) files="api/doc/swagger/$OPTARG.json";;
                \?) echo "Illegal option(s)"
                    exit;;
        esac
done

cnt=0

# Process files in api/http
for file in ./api/http/*.yaml; do
        if [ -f "$file" ]; then
            f=$(basename $file .yaml)
            mkdir -p ./api/doc/swagger

            api-spec-converter \
                    --from=openapi_3 \
                    --to=swagger_2 \
                    api/http/$f.yaml > api/doc/swagger/$f.json

            files="$files api/doc/swagger/$f.json"
            cnt=$((cnt+1))
        fi
done

# Process files in api/doc/swagger
for file in ./api/doc/swagger/*.yaml; do
        if [ -f "$file" ]; then
            f=$(basename $file .yaml)
            
            # Skip if JSON version already exists from previous step
            if [ ! -f "api/doc/swagger/$f.json" ]; then
                api-spec-converter \
                        --from=openapi_3 \
                        --to=swagger_2 \
                        api/doc/swagger/$f.yaml > api/doc/swagger/$f.json

                files="$files api/doc/swagger/$f.json"
                cnt=$((cnt+1))
            fi
        fi
done

# Only proceed if we have files to process
if [ $cnt -eq 0 ]; then
    echo "No OpenAPI files found. Exiting."
    exit 1
fi

mkdir -p ./docs/swagger
swagger -q mixin $files -o docs/swagger/docs.json 

echo "Swagger documentation generated successfully at docs/swagger/docs.json" 