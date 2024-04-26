# VreeDB ((V)ector-T(R)ee Database)
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
- persistance ON by default in an append-only file style

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Please use Docker to compile / install VreeDB:
- git clone https://github.com/spookieoli/VreeDB
- docker build -t vreedb .
- docker run -p 8080:8080 vreedb

Once up - you can access the UI at http://127.0.0.1:8080/
(WARNING! VREEDB IS IN EARLY ALPHA - DONT USE IT IN PRODUCTION!)

For API documentation please look at our Wiki: https://github.com/spookieoli/VreeDB/wiki
