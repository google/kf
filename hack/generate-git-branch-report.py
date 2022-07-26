#!/usr/bin/env python3

# Copyright 2020 Google LLC
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

import subprocess
import json


def git(command):
    call = subprocess.run(["git"] + command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8").strip()


def trim_name(name):
    origin_prefix = "origin/"
    if name.startswith(origin_prefix):
        return name[len(origin_prefix):]
    return name


def get_sha(name):
    return git("rev-parse %s" % name)


def get_email(name):
    return git("show -s --format=%ae " + name)


def get_date(name):
    return git("show -s --format=%ci " + name)


def get_title(name):
    lines = git("log --format=%B -n 1 " + name).split("\n")
    if len(lines) == 0:
        return ""
    return lines[0]


def parse_change_id(line):
    change_id_prefix = "Change-Id: "
    if line.startswith(change_id_prefix):
        return line[len(change_id_prefix):]


def get_change_id(name):
    for line in git("log --format=%B -n 1 " + name).split("\n"):
        change_id = parse_change_id(line)
        if change_id:
            return change_id
    return ""


def get_develop_logs():
    for line in git("log --format=%B remotes/origin/develop").split("\n"):
        yield line


def get_develop_change_ids():
    for line in get_develop_logs():
        change_id = parse_change_id(line)
        if change_id:
            yield change_id


head_change_ids = set()


def get_merged(change_id):
    if change_id in head_change_ids:
        return True

    for cid in get_develop_change_ids():
        head_change_ids.add(cid)
        if cid == change_id:
            return True

    return False


class Branch(dict):
    def __init__(self, name):
        self.name = trim_name(name)
        self.sha = get_sha(name)
        self.email = get_email(name)
        self.date = get_date(name)
        self.title = get_title(name)
        self.change_id = get_change_id(name)
        self.merged = get_merged(self.change_id)

        # We have to inherit from dict so we can be marshaled into JSON.
        dict.__init__(
            self,
            name=self.name,
            sha=self.sha,
            email=self.email,
            date=self.date,
            title=self.title,
            change_id=self.change_id,
            merged=self.merged
        )


def branches():
    denylisted_branches = {"", "HEAD", "develop", "master"}
    for name in git("for-each-ref --format=%(refname:short) refs/remotes/origin").split("\n"):
        b = Branch(name)
        if b.name in denylisted_branches:
            continue
        yield b


print(json.dumps(list(branches())))
