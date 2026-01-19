#!/usr/bin/env python3

# Copyright 2021 Google LLC
# Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script does the following things:
# - Enforce several rules (e.g. required fields) in metadata.yaml for each curated function.
# - Ensure the URLs in the examples are correct.
#
# To skip process a code block in a markdown file, you can add <!-- @skip -->
# before the code block. For example:
# <!-- @skip -->
# ```
# something here...
# ```

import json
import os
import subprocess

import yaml

metadata_filename = 'metadata.yaml'
directories_to_skip = ['_template', 'dist', 'node_modules']
examples_directory = 'examples'
functions_directory = 'functions'
lang_dirs = ['go', 'ts']
examples_directories_to_skip = ['_template']
fn_directories_to_skip = ['bind', 'inflate-helm-chart', 'set-gcp-resource-ids', 'set-name-prefix']
required_fields = ['image', 'description', 'tags', 'sourceURL', 'examplePackageURLs', 'emails', 'license']
kpt_team_email = 'kpt-team@google.com'
disallowed_kpt_commands = ['kpt fn run', 'kpt cfg', 'kpt pkg cat']
ghcr_prefix = 'ghcr.io/kptdev/krm-functions-catalog/'
ghcr_contrib_prefix = 'ghcr.io/kptdev/krm-functions-catalog/krm-fn-contrib/'
git_url_prefix = 'https://github.com/kptdev/krm-functions-catalog.git'
test_config_filename = 'config.yaml'
exec_script_filename = 'exec.sh'


def validate_main_branch(contrib=False):
    fn_name_to_examples = validate_functions_dir_for_main_branch(contrib)
    validate_examples_dir_for_main_branch(fn_name_to_examples, contrib)


def validate_examples_dir_for_main_branch(fn_name_to_examples, contrib):
    curr_examples_dir = examples_directory
    if contrib:
        curr_examples_dir = os.path.join('contrib', examples_directory)
    examples_seen = set()
    for fn_name, examples in fn_name_to_examples.items():
        for example_name in examples:
            examples_seen.add(example_name)
            if not example_name.startswith(fn_name):
                raise Exception(f'example name {example_name} must start with the function name {fn_name}')
            if example_name in examples_directories_to_skip:
                continue
            validate_example_md(fn_name, curr_examples_dir, example_name, 'main')
            if not eval_or_exec_script(os.path.join(curr_examples_dir, example_name)):
                validate_example_kptfile(fn_name, curr_examples_dir, example_name, 'main', contrib)

    for dir in os.listdir(curr_examples_dir):
        dir_name = os.path.join(curr_examples_dir, dir)
        if os.path.isdir(dir_name):
            if dir in examples_directories_to_skip:
                continue
            if not dir in examples_seen:
                raise Exception(f'directory {dir} is NOT in the metadata.yaml file of any functions')


def validate_functions_dir_for_main_branch(contrib):
    fn_name_to_examples = {}
    curr_fns_dir = functions_directory
    if contrib:
        curr_fns_dir = os.path.join('contrib', functions_directory)
    for dir_to_scan in [os.path.join(curr_fns_dir, lang) for lang in lang_dirs]:
        for file in os.listdir(dir_to_scan):
            if file in fn_directories_to_skip:
                continue
            path_name = os.path.join(dir_to_scan, file)
            if os.path.isdir(path_name):
                if file in directories_to_skip:
                    continue
                fn_name = file
                print(fn_name)
                if metadata_filename not in os.listdir(path_name):
                    raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
                meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
                example_list = [eg.split('/')[-1] for eg in meta['examplePackageURLs']]
                fn_name_to_examples[fn_name] = example_list
                validate_metadata(meta, 'main', path_name, fn_name, example_list, contrib)
    return fn_name_to_examples


def validate_release_branch(branch_name):
    items = branch_name.split('/')
    if len(items) != 2:
        raise Exception(f'the release branch name must be in the format of <fn-name>/v<major>.<minor>')
    fn_name = items[0]
    contrib, lang = contrib_and_lang(fn_name)
    examples = validate_functions_dir_for_release_branch(lang, branch_name, fn_name, contrib)
    validate_examples_dir_for_release_branch(branch_name, fn_name, examples, contrib)


def contrib_and_lang(fn_name):
    for lang in lang_dirs:
        if fn_name in os.listdir(os.path.join(functions_directory, lang)):
            return False, lang
        elif fn_name in os.listdir(os.path.join('contrib', functions_directory, lang)):
            return True, lang
    raise Exception(f"can't determine if {fn_name} belongs contrib catalog or its language")


def validate_examples_dir_for_release_branch(branch_name, fn_name, examples, contrib):
    curr_examples_dir = examples_directory
    if contrib:
        curr_examples_dir = os.path.join('contrib', examples_directory)
    for example_name in examples:
        validate_example_md(fn_name, curr_examples_dir, example_name, branch_name)
        if not eval_or_exec_script(os.path.join(curr_examples_dir, example_name)):
            validate_example_kptfile(fn_name, curr_examples_dir, example_name, branch_name, contrib)


def validate_functions_dir_for_release_branch(lang, branch_name, fn_name, contrib):
    print(f'verifying {fn_name}')
    path_name = os.path.join(functions_directory, lang, fn_name)
    if contrib:
        path_name = os.path.join('contrib', path_name)
    if metadata_filename not in os.listdir(path_name):
        raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
    meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
    examples = [eg.split('/')[-1] for eg in meta['examplePackageURLs']]
    validate_metadata(meta, branch_name, path_name, fn_name, examples, contrib)
    return examples


def eval_or_exec_script(example_path):
    test_config_filepath = os.path.join(example_path, '.expected', test_config_filename)
    if os.path.exists(test_config_filepath):
        test_config_file = yaml.load(open(test_config_filepath), Loader=yaml.Loader)
        if 'testType' in test_config_file and test_config_file['testType'] == 'eval':
            return True
    exec_script_filepath = os.path.join(example_path, '.expected', exec_script_filename)
    if os.path.exists(exec_script_filepath):
        return True
    return False


def latest_patch(fn_name, minor_version):
    scripts_dir = os.path.dirname(os.path.abspath(__file__))
    patch_reader = os.path.join(scripts_dir, 'patch_reader', 'patch_reader')
    process = subprocess.Popen([patch_reader,
                                '--function', fn_name,
                                '--minor', minor_version],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    if process.returncode != 0:
        raise Exception(f'patch_reader error: {stderr.decode("utf-8")}')
    patch_version = json.loads(stdout)
    return patch_version['latest_patch']


def validate_example_kptfile(fn_name, dir_name, example_name, branch, contrib):
    example_path = os.path.join(dir_name, example_name)
    kptfile_path = os.path.join(example_path, 'Kptfile')
    if not os.path.exists(kptfile_path):
        return

    if contrib:
        desired_ghcr_prefix = ghcr_contrib_prefix
    else:
        desired_ghcr_prefix = ghcr_prefix
    tag = branch
    if branch == 'main':
        tag = 'latest'
    else:
        splits = branch.split('/')
        if len(splits) != 2:
            raise Exception(f'the release branch {branch} must has format <fn-name>/vX.Y')
        minor_version = splits[1]
        tag = latest_patch(fn_name, minor_version)

    kptfile = yaml.load(open(kptfile_path), Loader=yaml.Loader)
    if kptfile['apiVersion'] != 'kpt.dev/v1alpha2' and kptfile['apiVersion'] != 'kpt.dev/v1':
        return

    # Stop processing when there are no pipeline declared.
    if 'pipeline' not in kptfile:
        return

    pipeline = kptfile['pipeline']
    if 'mutators' in pipeline:
        for mutator in pipeline['mutators']:
            actual_image = mutator['image']
            desired_image_name = f'{desired_ghcr_prefix}{fn_name}:{tag}'
            if actual_image != desired_image_name:
                raise Exception(f'expect Kptfile to contain {desired_image_name} but find {actual_image}')
    if 'validators' in pipeline:
        for mutator in pipeline['validators']:
            actual_image = mutator['image']
            desired_image_name = f'{desired_ghcr_prefix}{fn_name}:{tag}'
            if actual_image != desired_image_name:
                raise Exception(f'expect Kptfile to contain {desired_image_name} but find {actual_image}')


def validate_example_md(fn_name, dir_name, example_name, branch):
    example_path = os.path.join(dir_name, example_name)
    md_file_path = os.path.join(example_path, 'README.md')

    if (fn_name + "-") not in example_name:
        raise Exception(f'example directory "{example_name}" must have the function name "{fn_name}-" as prefix')

    with open(md_file_path) as f:
        first_line = f.readline().strip()
        if not first_line.startswith('# '):
            raise Exception(f'title must be in the 1st line and starts with one "#"')
        if fn_name not in first_line:
            raise Exception(f'title "{first_line}" must be in format "<fn-name>: Example Name" and contains "{fn_name}"')
        shorter_example_name = example_name
        if example_name.startswith(fn_name):
            shorter_example_name = example_name[len(fn_name):]
        if shorter_example_name.replace('-', ' ') not in first_line.lower():
            raise Exception(f'title "{first_line}" must be in format "<fn-name>: Example Name" and contains "{shorter_example_name.replace("-", " ")}"')

    process = subprocess.Popen(['mdrip', md_file_path],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    if len(stderr) > 0:
        print(f'stderr of mdrip: {str(stderr)}')

    process = subprocess.Popen(['mdrip', '--label', 'skip', md_file_path],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE)
    stdout2, stderr2 = process.communicate()
    if len(stderr2) > 0:
        print(f'stderr of mdrip: {str(stderr2)}')

    tag = branch
    if branch == 'main':
        tag = 'latest'
    else:
        splits = branch.split('/')
        if len(splits) != 2:
            raise Exception(f'the release branch {branch} must has format <fn-name>/vX.Y')
        tag = splits[1]

    lines = stdout.decode("utf-8").splitlines()
    lines2 = stdout2.decode("utf-8").splitlines()

    for line in lines:
        if line.startswith('#') or line.startswith('echo'):
            continue
        for disallowed in disallowed_kpt_commands:
            if disallowed in line:
                raise Exception(f'command {disallowed} is not allowed in the desired package url in {md_file_path}')

        if line in lines2:
            continue

        for item in line.split():
            if item.startswith(git_url_prefix):
                desired_pkg_url = f'{git_url_prefix}/{example_path}'
                if branch != 'main':
                    desired_pkg_url = desired_pkg_url + f'@{branch}'
                if not item.startswith(desired_pkg_url):
                    raise Exception(f'the desired package url in {md_file_path} is {desired_pkg_url}, but found {item}')

            if ghcr_prefix in item:
                desired_image_name = f'{ghcr_prefix}{fn_name}:{tag}'
                if desired_image_name not in item:
                    raise Exception(f'expect "{line}" to contain "{desired_image_name}" in {md_file_path}')


def validate_metadata(metadata, branch, path, fn, examples_list, contrib):
    for required_field in required_fields:
        if required_field not in metadata:
            raise Exception(f'{fn}: field {required_field} is required')
    if contrib:
        desired_image_name = ghcr_contrib_prefix + fn
    else:
        desired_image_name = ghcr_prefix + fn
    if metadata['image'] != desired_image_name:
        raise Exception(f'{fn}: image name should be "{desired_image_name}"')
    if metadata['tags'] is None or len(metadata['tags']) == 0:
        raise Exception(f'{fn}: "tags" must contain at least one tag')
    desired_source_url = f'https://github.com/kptdev/krm-functions-catalog/tree/{branch}/{path}'
    if metadata['sourceURL'] != desired_source_url:
        raise Exception(f'{fn}: "sourceURL" should be "{desired_source_url}"')
    if len(examples_list) == 0:
        raise Exception(f'{fn}: there must be at least one example listed in metadata.yaml file')
    for example in examples_list:
        if contrib:
            desired_example_pkg_url = f'https://github.com/kptdev/krm-functions-catalog/tree/{branch}/contrib/examples/{example}'
        else:
            desired_example_pkg_url = f'https://github.com/kptdev/krm-functions-catalog/tree/{branch}/examples/{example}'
        if desired_example_pkg_url not in metadata['examplePackageURLs']:
            raise Exception(f'"{desired_example_pkg_url}" is not listed in examplePackageURLs')
    if metadata['emails'] is None or kpt_team_email not in metadata['emails']:
        raise Exception(f'"{kpt_team_email}" should be in the emails list')


def main():
    branch_name = os.getenv('GITHUB_BASE_REF')
    if branch_name is None or len(branch_name) == 0:
        branch_name = 'main'

    if branch_name == 'main':
        print("verifying curated functions catalog...")
        validate_main_branch()
        print("verifying contrib functions catalog...")
        validate_main_branch(contrib=True)
    else:
        validate_release_branch(branch_name)
    print("Docs verification succeeded!")


if __name__ == "__main__":
    main()
