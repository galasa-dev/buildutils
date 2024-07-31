#! /usr/bin/env bash 

#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
#-----------------------------------------------------------------------------------------                   
#
# Objectives: Build this repository code locally.
# 
#-----------------------------------------------------------------------------------------                   

# Where is this script executing from ?
BASEDIR=$(dirname "$0");pushd $BASEDIR 2>&1 >> /dev/null ;BASEDIR=$(pwd);popd 2>&1 >> /dev/null
# echo "Running from directory ${BASEDIR}"
export ORIGINAL_DIR=$(pwd)
# cd "${BASEDIR}"

cd "${BASEDIR}/.."
WORKSPACE_DIR=$(pwd)


#-----------------------------------------------------------------------------------------                   
#
# Set Colors
#
#-----------------------------------------------------------------------------------------                   
bold=$(tput bold)
underline=$(tput sgr 0 1)
reset=$(tput sgr0)
red=$(tput setaf 1)
green=$(tput setaf 76)
white=$(tput setaf 7)
tan=$(tput setaf 202)
blue=$(tput setaf 25)

#-----------------------------------------------------------------------------------------                   
#
# Headers and Logging
#
#-----------------------------------------------------------------------------------------                   
underline() { printf "${underline}${bold}%s${reset}\n" "$@"
}
h1() { printf "\n${underline}${bold}${blue}%s${reset}\n" "$@"
}
h2() { printf "\n${underline}${bold}${white}%s${reset}\n" "$@"
}
debug() { printf "${white}[.] %s${reset}\n" "$@"
}
info()  { printf "${white}[➜] %s${reset}\n" "$@"
}
success() { printf "${white}[${green}✔${white}] ${green}%s${reset}\n" "$@"
}
error() { printf "${white}[${red}✖${white}] ${red}%s${reset}\n" "$@"
}
warn() { printf "${white}[${tan}➜${white}] ${tan}%s${reset}\n" "$@"
}
bold() { printf "${bold}%s${reset}\n" "$@"
}
note() { printf "\n${underline}${bold}${blue}Note:${reset} ${blue}%s${reset}\n" "$@"
}

#-----------------------------------------------------------------------------------------                   
# Functions
#-----------------------------------------------------------------------------------------                   
function usage {
    info "Syntax: build-locally.sh [OPTIONS]"
    cat << EOF
Options are:
-h | --help : Display this help text
EOF
}

function check_exit_code () {
    # This function takes 2 parameters in the form:
    # $1 an integer value of the returned exit code
    # $2 an error message to display if $1 is not equal to 0
    if [[ "$1" != "0" ]]; then 
        error "$2" 
        exit 1  
    fi
}

function check_secrets {
    h2 "updating secrets baseline"
    cd ${BASEDIR}
    detect-secrets scan --exclude-files "go.sum|openapi2beans/go.sum" --update .secrets.baseline
    rc=$? 
    check_exit_code $rc "Failed to run detect-secrets. Please check it is installed properly" 
    success "updated secrets file"

    h2 "running audit for secrets"
    detect-secrets audit .secrets.baseline
    rc=$? 
    check_exit_code $rc "Failed to audit detect-secrets."
    
    #Check all secrets have been audited
    secrets=$(grep -c hashed_secret .secrets.baseline)
    audits=$(grep -c is_secret .secrets.baseline)
    if [[ "$secrets" != "$audits" ]]; then 
        error "Not all secrets found have been audited"
        exit 1  
    fi
    sed -i '' '/[ ]*"generated_at": ".*",/d' .secrets.baseline
    success "secrets audit complete"
}

#-----------------------------------------------------------------------------------------                   
# Process parameters
#-----------------------------------------------------------------------------------------                   
exportbuild_type=""

while [ "$1" != "" ]; do
    case $1 in
        -h | --help )           usage
                                exit
                                ;;
        * )                     error "Unexpected argument $1"
                                usage
                                exit 1
    esac
    shift
done

#-----------------------------------------------------------------------------------------                   
# Main logic.
#-----------------------------------------------------------------------------------------                   

source_dir="."

function clean_temp_folder() {
    rm -fr $BASEDIR/temp
    mkdir -p $BASEDIR/temp
    LOGS_DIR=$BASEDIR/temp
}

function build_tools() {

    project=$(basename ${BASEDIR})
    h1 "Building ${project}"

    info "Using source code at ${source_dir}"
    cd ${BASEDIR}/${source_dir}

    log_file=${LOGS_DIR}/${project}.txt
    info "Log will be placed at ${log_file}"
    date > ${log_file}

    cmd="make all"
    info "Command is '$cmd'"

    cd ${BASEDIR}
    $cmd 2>&1 >> ${log_file}

    rc=$? 
    check_exit_code $rc "Failed to build the ${project}" 
    success "${project} built ok - log is at ${log_file}"
}

function setup_source_folder() {
    rm -fr  $BASEDIR/temp/src
    mkdir -p $BASEDIR/temp/src

    mkdir -p $BASEDIR/temp/src/dev/galasa/examples/module1
    cat << EOF > $BASEDIR/temp/src/dev/galasa/examples/module1/build.gradle
# A test module mock-up.
version = "0.0.1-SNAPSHOT" // trailing comment
# trailing content
EOF

    cat << EOF > $BASEDIR/temp/src/dev/galasa/examples/module1/settings.gradle
# initial content
rootProject.name = "dev.galasa.examples/module1"
# trailing content
EOF

    mkdir -p $BASEDIR/temp/src/dev/galasa/examples/module2
    cat << EOF > $BASEDIR/temp/src/dev/galasa/examples/module2/build.gradle
# A test module mock-up.
version = "0.0.2-SNAPSHOT" // trailing comment
# trailing content
EOF

    cat << EOF > $BASEDIR/temp/src/dev/galasa/examples/module2/settings.gradle
# initial content
rootProject.name = "dev.galasa.examples/module2"
# trailing content
EOF

}

function check_versions_have_suffix() {
    suffix=$1
    if [[ "$suffix" == "" ]]; then 
        info "Checking that the versions of the code are not using a suffix"
    else
        info "Checking that the versions of the code are using the $suffix suffix"
    fi

    cat << EOF > $BASEDIR/temp/versions-list-expected.txt
[$GALASABLD versioning list --sourcefolderpath $BASEDIR/temp/src]
dev.galasa.examples/module1 0.0.1$suffix
dev.galasa.examples/module2 0.0.2$suffix
EOF

    diff $BASEDIR/temp/versions-list-expected.txt $BASEDIR/temp/versions-list-got.txt >> /dev/null
    rc=$? ; if [[ "$rc" != "0" ]]; then error "Output from listing versions is not what we expected." ; exit 1 ; fi
    success "The list of versions is what we expected."
}

function clear_version_suffixes() {
    info "Removing the suffixes"
    cmd="$GALASABLD versioning suffix remove --sourcefolderpath $BASEDIR/temp/src "

    info "Command is $cmd"
    $cmd > $BASEDIR/temp/versions-removed.txt
    rc=$? ; if [[ "$rc" != "0" ]]; then error "Could not remove the version suffixes of the code. rc=$?" ; exit 1 ; fi
    success "Version suffixes of modules removed OK."
}

function gather_version_list() {
    info "Listing the suffixes"
    cmd="$GALASABLD versioning list --sourcefolderpath $BASEDIR/temp/src "

    info "Command is $cmd"
    $cmd > $BASEDIR/temp/versions-list-got.txt
    rc=$? ; if [[ "$rc" != "0" ]]; then error "Could not set the versions of the code. rc=$?" ; exit 1 ; fi
    success "Versions of modules set OK."
}

function set_version_suffixes() {
    desired_suffix=$1
    info "Setting the suffixes prefixes to $desired_suffix"
    cmd="$GALASABLD versioning suffix set --sourcefolderpath $BASEDIR/temp/src --suffix $desired_suffix"

    info "Command is $cmd"
    $cmd > $BASEDIR/temp/versions-set.txt
    rc=$? ; if [[ "$rc" != "0" ]]; then error "Could not set the version suffixes to $desired_suffix. rc=$?" ; exit 1 ; fi
    success "Version suffixes of modules set to $desired_suffix OK."
}

function test_versions_manipulation() {
    h2 "Testing manipulations of versions"

    setup_source_folder
    gather_version_list
    check_versions_have_suffix "-SNAPSHOT"

    info "Removing the suffixes on versions"
    clear_version_suffixes
    gather_version_list
    check_versions_have_suffix ""

    info "Setting the suffixes on versions"
    set_version_suffixes "-alpha"
    gather_version_list
    check_versions_have_suffix "-alpha"

    success "Tested the galasabld versioning commands as best we can"
}

function build_openapi2beans() {
    h2 "Building openapi2beans."
    ./openapi2beans/build-locally.sh
}

clean_temp_folder
build_tools

export GALASABLD=${BASEDIR}/bin/galasabld-darwin-arm64

test_versions_manipulation
build_openapi2beans
check_secrets