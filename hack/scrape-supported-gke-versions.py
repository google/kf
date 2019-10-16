#!/usr/bin/env python3

# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the License);
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an AS IS BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script uses the chrome web driver and selenium to scapre the GCP docs for
# supported Cloud Run versions on GKE
#
# Prerequisites for this script are:
# - pip install selenium
# - download chromedriver from https://chromedriver.chromium.org/downloads

from selenium import webdriver
import json

options = webdriver.ChromeOptions()
options.add_argument('headless')
browser = webdriver.Chrome(options=options)
browser.get("https://cloud.google.com/run/docs/gke/cluster-versions")

# assume there's only one div with this class.
elem = browser.find_element_by_class_name("devsite-table-wrapper")
table = elem.find_element_by_tag_name("table")
rows = table.find_elements_by_tag_name("tr")

title = True
supported_versions = []
for row in rows:
    if title:
        title = False
        continue
    cols = row.find_elements_by_tag_name("td")
    cloud_run_cell = cols[0]
    gke_cell = cols[1]
    cloud_run_version = cloud_run_cell.text
    gke_verions = gke_cell.find_elements_by_tag_name("p")
    gke_verions = [p.text for p in gke_verions]
    supported_versions.append({
        "cloud_run_version": cloud_run_version,
        "gke_versions": gke_verions,
    })

print(json.dumps(supported_versions))
