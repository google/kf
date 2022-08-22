---
title: Get started with buildpacks
description: "How to use built-in buildpacks for various languages."
weight: 200
---

Kf supports a variety of buildpacks. This document
covers some starter examples for using them.

## Before you begin

* You should have Kf running on a cluster.
* You should have run `kf target -s <space-name>` to target your space.

## Java (v2) buildpack

Use [spring initializr](https://start.spring.io/) to create a Java 8 maven project with a spring web dependency and JAR packaging. Download it, extract it, and once extracted you can generate a JAR.

```sh
./mvnw package
```

Push the JAR to Kf with the Java v2 buildpack.

```sh
kf push java-v2 --path target/helloworld-0.0.1-SNAPSHOT.jar
```

## Java (v3) buildpack

Use [spring initializr](https://start.spring.io/) to create a Java 8 maven project with a spring web dependency and JAR packaging. Download it, extract it, and once extracted, push to Kf with the cloud native buildpack.

```sh
kf push java-v3 --stack org.cloudfoundry.stacks.cflinuxfs3
```

## Python (v2) buildpack

Create a new directory with files as shown in the following structure.

```sh
tree
.
├── Procfile
├── requirements.txt
└── server.py
```

```sh
cat Procfile
web: python server.py
```

```sh
cat requirements.txt
Flask
```

```sh
cat server.py
from flask import Flask
import os

app = Flask(__name__)

@app.route('/')
def hello_world():
    return 'Hello, World!'

if __name__ == "__main__":
  port = int(os.getenv("PORT", 8080))
  app.run(host='0.0.0.0', port=port)
```

Push the Python flask app using v2 buildpacks.

```sh
kf push python --buildpack python\_buildpack
```

## Python (v3) buildpack

(same as above)

Push the Python flask app using cloud native buildpacks.

```sh
kf push pythonv3 --stack org.cloudfoundry.stacks.cflinuxfs3
```

## Staticfile (v2) buildpack

Create a new directory that holds your source code.

Add an `index.html` file with this content.

```html
<!DOCTYPE html>

<html lang="en">

<head><title>Hello, world!</title></head>

<body><h1>Hello, world!</h1></body>

</html>
```

Push the static content with the staticfile buildpack.

```sh
kf push staticsite --buildpack staticfile\_buildpack
```
