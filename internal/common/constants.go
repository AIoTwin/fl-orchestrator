package common

// Container images
const FL_CLIENT_IMAGE = "aiotwin/hfl-client:0.1"
const LOCAL_AGGRETATOR_IMAGE = "cilicivan96/hfl-local-server:0.2"
const GLOBAL_AGGRETATOR_IMAGE = "aiotwin/hfl-global-server:0.1"

// FL Client configs
const FL_CLIENT_DEPLOYMENT_PREFIX = "fl-cl"
const FL_CLIENT_CONFIG_MOUNT_PATH = "/home/"
const FL_CLIENT_CONFIG_MAP_NAME = "fl-cl-cm"
const FL_CLIENT_CPU_REQUEST = 0.5

// GA configs
const GLOBAL_AGGRETATOR_DEPLOYMENT_NAME = "fl-ga"
const GLOBAL_AGGRETATOR_MOUNT_PATH = "/home/"
const GLOBAL_AGGREGATOR_SERVICE_NAME = "fl-ga-svc"
const GLOBAL_AGGREGATOR_CONFIG_MAP_NAME = "fl-ga-cm"

const GLOBAL_AGGREGATOR_PORT = 8080
const GLOBAL_AGGREGATOR_ROUNDS = 100

// LA configs
const LOCAL_AGGRETATOR_DEPLOYMENT_PREFIX = "fl-la"
const LOCAL_AGGRETATOR_MOUNT_PATH = "/home/"
const LOCAL_AGGREGATOR_SERVICE_NAME = "fl-la-svc"
const LOCAL_AGGREGATOR_CONFIG_MAP_NAME = "fl-la-cm"

const LOCAL_AGGREGATOR_PORT = 8080
const LOCAL_AGGREGATOR_ROUNDS = 100

// FL types
const FL_TYPE_CLIENT = "client"
const FL_TYPE_LOCAL_AGGREGATOR = "local_aggregator"
const FL_TYPE_GLOBAL_AGGREGATOR = "global_aggregator"

// Events
const NODE_STATE_CHANGE_EVENT_TYPE = "NodeStateChanged"
const FL_FINISHED_EVENT_TYPE = "FlFinished"

// Node states
const NODE_ADDED = "ADDED"
const NODE_REMOVED = "REMOVED"
