module default {
    type User {
        required property email -> str;
        required property password_hash -> str;
        required property first_name -> str;
        required property last_name -> str;
        required property user_type -> str;
        required property user_id -> str;
        required property created_at -> datetime;
        required property updated_at -> datetime;
        required property token -> str;
        required property refresh_token -> str;
    }
};

