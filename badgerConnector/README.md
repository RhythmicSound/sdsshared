# badgerConnector

## Requirements
Datasets using Badger must:
- Have a key `_version` in the database which holds a JSON encoded sdsshared.VersionManager object. This value will be displayed in json response under `meta` in the form:
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