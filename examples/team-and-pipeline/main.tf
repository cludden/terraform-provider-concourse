provider "concourse" {
  username      = "test"
  password      = "test"
  concourse_url = "http://localhost:8080"
}

resource "concourse_team" "avengers" {
  name       = "avengers"
  auth_users = ["cludden"]
}

resource "concourse_pipeline" "batman" {
  team   = concourse_team.avengers.name
  name   = "batman"
  config = <<PIPELINE
jobs:
  - name: job
    public: true
    plan:
      - task: simple-task
        config:
          platform: linux
          image_resource:
            type: registry-image
            source: { repository: busybox }
          run:
            path: echo
            args: ["Hello, world!"]
PIPELINE
}