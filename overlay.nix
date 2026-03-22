# Knockout overlay — evaluated by fort-overlay-manager at activation time.
# Arguments: storePath (injected by manager), port (from host config).
{ port, storePath, ... }:
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
    ];
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
