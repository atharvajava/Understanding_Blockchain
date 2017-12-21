import hashlib,json,sys

def hashMe(msg=""):
    #Helper function to implement hashing alogorithm
    if type(msg)!=str:
        msg=json.dumps(msg,sort_keys=True)
        #if we dont sort keys we cannot guarantee repeatability
        #ok now i don't understand this part
    if sys.version_info.major == 2:
        return unicode(hashlib.sha256(msg).hexdigest(),'utf-8')
    else:
        return hashlib.sha256(str(msg).encode('utf-8')).hexdigest()

def checkBlockHash(block):
    # Raise an exception if the hash does not match the block contents
    expectedHash = hashMe( block['contents'] )
    if block['hash']!=expectedHash:
        raise Exception('Hash does not match contents of block %s'%
                        block['contents']['blockNumber'])
    return

