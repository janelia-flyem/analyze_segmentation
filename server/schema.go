package server

const graphSchema = `
{ "$schema": "http://json-schema.org/schema#",
  "title": "Define graph and connection uncertainties",
  "type": "object",
  "properties": {
    "edge_list": { 
      "description": "list of edges making up graph",
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "node1": {
            "description": "id of edge node 1",
            "type": "number",
            "minimum": 1
          },
          "node2": {
            "description": "id of edge node 2",
            "type": "number",
            "minimum": 1
          },
          "size1": {
            "description": "size of edge node 1",
            "type": "number",
            "minimum": 0
          },
          "size2": {
            "description": "size of edge node 2",
            "type": "number",
            "minimum": 0
          },
          "weight": {
            "description": "edge uncertainty (0: confident merge, 1: confident split)",
            "type": "number",
            "minimum": 0
          }
        },
        "required" : ["node1", "node2", "size1", "size2", "weight"]
      }
    }
  },
  "required" : ["edge_list"]
}


`
