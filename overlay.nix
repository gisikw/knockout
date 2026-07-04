# Knockout overlay — evaluated by fort-overlay-manager at activation time.
# Arguments: storePath (injected by manager), port (from host config).
#
# QQL cutover env (all optional, empty = inert, so this overlay behaves exactly
# like the pre-cutover one until the host sets these in overlays.<host>.config):
#   koQql        -> KO_QQL          (1 routes ko serve to questbook's QQL API)
#   koQqlUrl     -> KO_QQL_URL      (questbook base URL, e.g. loopback overlay)
#   koQqlMapping -> KO_QQL_MAPPING  (project->realm mapping YAML path)
#   koReadonly   -> KO_READONLY     (1 rejects writes to the legacy store)
#   koShimLog    -> KO_SHIM_LOG     (serve-side JSONL usage log — cutover vital sign)
{ port, storePath, koQql ? "", koQqlUrl ? "", koQqlMapping ? "", koReadonly ? ""
, koShimLog ? "", ... }:
let
  optEnv = name: val: if val == "" then [ ] else [ "${name}=${val}" ];
in
{
  services.knockout = {
    exec = "${storePath}/bin/ko serve --port ${port}";
    user = "dev";
    group = "users";
    workingDirectory = "/home/dev/Projects/exocortex";
    after = [ "network.target" ];
    restart = "on-failure";
    restartSec = 5;
    environment = [
      "PATH=${storePath}/bin:/run/current-system/sw/bin"
    ]
    ++ optEnv "KO_QQL" koQql
    ++ optEnv "KO_QQL_URL" koQqlUrl
    ++ optEnv "KO_QQL_MAPPING" koQqlMapping
    ++ optEnv "KO_READONLY" koReadonly
    ++ optEnv "KO_SHIM_LOG" koShimLog;
  };

  bins = [ "${storePath}/bin/ko" ];

  health = {
    type = "tcp";
    endpoint = "127.0.0.1:${port}";
    interval = 2;
    grace = 3;
    stabilize = 10;
  };
}
