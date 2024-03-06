package common

// Container images
const FL_CLIENT_IMAGE = "cilicivan96/aiotwin-fl-client:1.0"
const GLOBAL_AGGRETATOR_IMAGE = "cilicivan96/aiotwin-fl-global-server:1.0"

const FL_CLIENT_DEPLOYMENT_PREFIX = "fl-cl"
const FL_CLIENT_CONFIG_MOUNT_PATH = "/app/config/example_client/"
const FL_CLIENT_CONFIG_MAP_NAME = "fl-cl-cm"

const GLOBAL_AGGRETATOR_DEPLOYMENT_NAME = "fl-ga"
const GLOBAL_AGGRETATOR_MOUNT_PATH = "/app/config/example_global_server/"
const GLOBAL_AGGREGATOR_SERVICE_NAME = "fl-ga-svc"
const GLOBAL_AGGREGATOR_CONFIG_MAP_NAME = "fl-ga-cm"

const FL_TYPE_CLIENT = "client"
const FL_TYPE_AGGREGATOR = "aggregator"

const GLOBAL_AGGREGATOR_PORT = 8080
const GLOBAL_AGGREGATOR_ROUNDS = 100
