# External Interfaces

external interface related implementations are located here and under this directory

### extintf
This directory owns the implementation regarding external resources and interfaces.
In case of a external interface change, this directory expected to be affected only.

#### httpintf
This directory responsible for owning HTTP protocol based external interfaces,
such as API, WebGUI or monitoring integration.

##### httpapi
httpapi implements the API that allows clients to interact with

#### storages
Storages pkg act as main entry point for owning storage implementations.
It has a factory method that can return the proper implementation based on the connection string.

##### postgres
postgres fulfills the requirements made by the shared specifications.
Act as a db implementation to store/retrieve data.
