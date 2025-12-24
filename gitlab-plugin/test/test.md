# api


## get projects

```
curl --request GET --header "Authorization: Bearer kjzzDRV5ysCUY2AD_jrC" --url "https://gitlab-ce.alauda.cn/api/v4/projects"
```

**response**

[gitlab_project_response](./gitlab_project_response.json)

## create projects

```
curl --request POST --header "Authorization: Bearer kjzzDRV5ysCUY2AD_jrC" --header "Content-Type: application/json" --data '{"name": "hxia-project", "description": "New Project for gitlab", "path": "hxia-project","namespace": "Huyun Xia", "namespace_id": "1038", "initialize_with_readme": "true"}' --url "https://gitlab-ce.alauda.cn/api/v4/projects/"
```

**response**


## get members

```
curl --request GET --header "Authorization: Bearer kjzzDRV5ysCUY2AD_jrC" --url "https://gitlab-ce.alauda.cn/api/v4/projects/4538/members"
```

**response**

[gitlab_member_response](./gitlab_member_response.json)


