#!/usr/bin/env bash

set -eux

project_id=$1
version=$2

if [ -z "${project_id}" ] || [ -z "${version}" ]; then
  echo "usage: $0 [PROJECT_ID] [VERSION]"
  exit 1
fi

temp_dir=$(mktemp -d)
builder_config_file="$temp_dir/builder.toml"
buildpacks_file=$temp_dir/buildpacks.txt
echo $buildpacks_file
echo "version: $version, project_id: $project_id"
declare -a scripts=("detect" "compile" "release" "finalize", "supply")

cat << EOF >> $buildpacks_file
https://github.com/cloudfoundry/java-buildpack
https://github.com/cloudfoundry/dotnet-core-buildpack
https://github.com/cloudfoundry/nodejs-buildpack
https://github.com/cloudfoundry/go-buildpack
https://github.com/cloudfoundry/python-buildpack
https://github.com/cloudfoundry/binary-buildpack
https://github.com/cloudfoundry/nginx-buildpack
EOF

function create_buildpack_images {
    while read bp_url; do
    bp_name="${bp_url##*/}"
    bp_name_underscores=$(echo $bp_name | sed 's|-|_|g')

    output_dir="$temp_dir/$bp_name/bin"
    mkdir -p $output_dir
        # Create scripts
        for script in "${scripts[@]}"
        do
            cat << EOF >> $output_dir/$script
#!/usr/bin/env bash
set -euo pipefail

if [ ! -d "/tmp/git-$bp_name" ]; then
    git clone $bp_url /tmp/git-$bp_name
fi
exec /tmp/git-$bp_name/bin/$script "\$@"
EOF
            chmod 777 $output_dir/$script
        done

        # Create builder image with `kf wrap-v2-buildpack`
        bp_dir="$temp_dir/$bp_name"
        /workspace/bin/kf-linux wrap-v2-buildpack $bp_name_underscores $bp_dir --buildpack-version $version --buildpack-stacks io.buildpacks.stacks.bionic
    done <$buildpacks_file
}
create_buildpack_images

function create_builder_toml {
    while read bp_url; do
        bp_name="${bp_url##*/}"
        bp_name_underscores=$(echo $bp_name | sed 's|-|_|g')
        bp_uri="$bp_name_underscores:$version"
        cat << EOF >> $builder_config_file
[[buildpacks]]
id = "$bp_name_underscores"
uri = "$bp_uri"
version = "$version"

EOF
    done <$buildpacks_file

    while read bp_url; do
        bp_name="${bp_url##*/}"
        bp_name_underscores=$(echo $bp_name | sed 's|-|_|g')
        cat << EOF >> $builder_config_file
[[order]]
   [[order.group]]
   id = "$bp_name_underscores"

EOF
    done <$buildpacks_file

            cat << EOF >> $builder_config_file
[stack]
build-image = "cloudfoundry/build:1.3.2-full-cnb"
id = "io.buildpacks.stacks.bionic"
run-image = "cloudfoundry/run:1.1.37-base-cnb"
EOF
}
create_builder_toml

function create_builder_image {
    builder_name="gcr.io/$project_id/v2-to-v3-builder:$version"
    pack builder create $builder_name --config $builder_config_file
}
create_builder_image
exit
