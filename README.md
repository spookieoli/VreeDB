# VreeDB ((V)ector-T(ree) Database)
A simple, fast, no dependency, Go powered Vector Database

VreeDB is a database project written in Go. It uses a k-d tree data structure for efficient search of nearest neighbors in a multi-dimensional space.

## Features

- Efficient nearest neighbor search using k-d trees.
- Easy to use
- Platform independent (when using Docker)
- Includes Classifiers (SVM / Neural Nets) - you can use the Database as Classifier (Pictures, Text, Sound etc etc)
- "Distanceaware" Filtering
- Supports multiple collections.
- Backup and restore functionality for collections.
- Thread-safe operations.
- Persistence ON by default in an append-only file style

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Please use git clone to copy the code in the repository from GitHub:

Install git from [git's official website](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) and follow the instructions given there.

Once this is done, create a folder called VreeDB in your local computer and navigate to that folder in a bash window or a command shell window. Then, execute the below command in the bash terminal or command shell that you've opened. 

```bash
git clone https://github.com/spookieoli/VreeDB
```

## Please use Docker to compile / install VreeDB:

Install docker from [docker's official website](https://docs.docker.com/engine/install/) and follow the instructions given there.

You would need to navigate to the home directory of the cloned repository locally in a bash window or command shell window and run the below command.

```bash
docker build -t vreedb .
```

If you need a simple and quick installation of the product, run the command give below in the same bash window or command shell window.

```bash
docker run -p 8080:8080 vreedb
```

If you would like to persist the data in the database across sessions and you are having a windows laptop then simply use the below docker run command instead of the earlier one:

```bash
docker run -v C:\collections:/collections -p 8080:8080 --name vreedb_test vreedb
```

Once up - you can access the UI by clicking [here](http://127.0.0.1:8080/) or by copy-pasting the URL below in the address bar of your favorite browser:

```
http://127.0.0.1:8080/
```


⚠️ **Warning:** VreeDB IS IN EARLY ALPHA - DON'T USE IT IN PRODUCTION!

👁️ **Sneak Peek:** Version 1.0 is just around the corner!

For API documentation please look at our Wiki: https://github.com/spookieoli/VreeDB/wiki
