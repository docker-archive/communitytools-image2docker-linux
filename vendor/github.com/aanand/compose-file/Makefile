SCHEMA_GO := schema/bindata.go
SCHEMA_JSON := schema/data/config_schema_v3.0.json

test:
	go test ./{loader,schema,template,interpolation}

schema: $(SCHEMA_GO)

$(SCHEMA_GO): $(SCHEMA_JSON)
	go generate ./schema
