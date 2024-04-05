# VreeDB
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

## Please use Docker to compile / install VectoriaDB (WARNING! VREEDB IS IN EARLY ALPHA - DONT USE IT IN PRODUCTION!):
- git clone https://github.com/spookieoli/VreeDB
- docker build -t vreedb .
- docker run -p 8080:8080 vreedb

### Documentation Outline

#### Overview
- Brief introduction to the VreeDB and its RESTful interface.
- General notes on authentication, error handling, and response formats.

#### Types Documentation

##### `CreateCollection`
- **Description**: Struct used to create a new Collection in the VDB.
- **REST Method**: `POST`
- **Endpoint**: `/createcollection`
- **Fields**:
  - `ApiKey` (string): Authentication key, may or may not included in the REST request.
  - `Name` (string): Name of the collection to create.
  - `DistanceFunction` (string): The function used to calculate the distance between vectors in the collection.
  - `Dimensions` (int): Number of dimensions for the vectors in the collection.
  - `Wait` (bool): If true, the request waits for the collection to be fully created before returning.

##### `Delete`
- **Description**: Struct used to delete an existing Collection from the VDB.
- **REST Method**: `DELETE`
- **Endpoint**: `/delete`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `Name` (string): Name of the collection to delete.

##### `ListCollections`
- **Description**: Struct for listing all Collections in the VDB.
- **REST Method**: `GET`
- **Endpoint**: `/ListCollections`
- **Fields**:

##### `AddPoint`
- **Description**: Struct for adding a point to a Collection.
- **REST Method**: `POST`
- **Endpoint**: `/AddPoint`
- **Fields**:
  - `Id` (string): Unique identifier for the point, may or may not included in the REST request.
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection to add the point to.
  - `Vector` ([]float64): The vector data.
  - `Payload` (map[string]interface{}): Optional metadata associated with the point.
  - `Depth`, `Wait`, `MaxDistancePercent`: Various parameters not included in the REST request, with default values.
 
##### `Search`
- **Description**: Struct for adding a point to a Collection.
- **REST Method**: `POST`
- **Endpoint**: `/Search`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection to add the point to.
  - `Vector` ([]float64): The vector data.

##### `AddPointBatch`
- **Description**: Struct for adding a batch of points to a Collection.
- **REST Method**: `POST`
- **Endpoint**: `/AddPointBatch`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection to add points to.
  - `Points` ([]PointItem): List of points to add.

##### `DeletePoint`
- **Description**: Struct for deleting a point from a Collection.
- **REST Method**: `DELETE`
- **Endpoint**: `/DeletePoint`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection to delete the point from.
  - `Id` (string): Unique identifier of the point to delete.
 
##### `TrainClassifier`
- **Description**: Struct for deleting a classifier.
- **REST Method**: `PUT`
- **Endpoint**: `/TrainClassifier`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection associated with the classifier.
  - `ClassifierName` (string): Name of the classifier to delete.

##### `DeleteClassifier`
- **Description**: Struct for deleting a classifier.
- **REST Method**: `DELETE`
- **Endpoint**: `/DeleteClassifier`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection associated with the classifier.
  - `ClassifierName` (string): Name of the classifier to delete.

##### `Classify`
- **Description**: Struct for classifying a vector.
- **REST Method**: `POST`
- **Endpoint**: `Classify`
- **Fields**:
  - `ApiKey` (string): Authentication key.
  - `CollectionName` (string): Name of the collection.
  - `ClassifierName` (string): Name of the classifier to use.
  - `Vector` ([]float64): The vector to classify.

#### Additional Information
- Example requests and responses for each endpoint. (TBD)
- Best practices for using the API. (TBD)
- Information on rate limiting, if applicable. (TBD)

