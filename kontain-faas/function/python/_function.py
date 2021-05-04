from kontain import snapshots
import json
import base64

data_in = snapshots.getdata()
print("data_in:", data_in)
in_struct = json.loads(data_in)

snapshots.putdata(json.dumps({ 'Status': 200, 'Data': base64.b64encode(b'Hello from python').decode()}))