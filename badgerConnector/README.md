# badgerConnector

## Requirements
Datasets using Badger must:
- Have a key `_version` in the database which holds a JSON encoded sdsshared.Meta object (which can be unmarshalled into a sdsshared.VersionManager or vice versa if needs be). This value will be displayed in json response under `meta` in the form:
```json
{
     "meta": {
       "resource": "Default Resource Name",
       "dataset_updated": "2021-12-08T21:07:33Z",
       "data_sources": [
        "https://osdatahub.os.uk/downloads/open#OPNAME"
       ]
     }
}
```
- Must have all other keys encoded using sds.CreateKVStoreKey and a seperator of `/` or a similar function that creates reliably unique composite keys of format "**`$lookup_value/$unique_identifier`**" so that even duplicate lookup values committed at differenct times can coexist and do not attempt to overwrite. This value will be displayed in json response in the form 
```json
{  
    "data": {
       "values": {
            "SE129TA": "1638997647343965567",
            "SE129TE": "1638997647344003584",
            "$lookup_value": "$unique_identifier",
   }
}
```
## Usage
BadgerConnect accepts an additional query key:Value `predict` as a bool for whether or not to return an autosuggest list of available keys (true) of a single return value (false).

So an example query to a url endpoint of a server implenting BadgerConnector is like: 
```markdown
http:/localhost:8080/fetch?q=CR05qp&predict=false
```