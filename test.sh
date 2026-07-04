PAT="mc_pat_f13c68d4379555434c2cdcf168b40754c4d89858913564ce46a2ee24dcac99aa"
BASE="http://localhost:8080/end-user/graphql/org/luke_e6kz/project/luke/db/demo_ecommerce/model/orders"

curl -X POST "$BASE" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $PAT" \
    -H "X-MC-Auth-Useadmin: true" \
	--data-raw '{"query":"query { findMany(take:5,skip:0,orderBy:[{id:asc}],where:{user_id:{equals:\"det-test-user-001\"}}) { items { id } totalCount } }"}'
