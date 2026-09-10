resource "gitea_repository_topics" "example" {
  user   = "my-org"
  repo   = "my-repo"
  topics = ["terraform", "gitea", "automation"]
}
