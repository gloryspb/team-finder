export type User = {
  id: string;
  email: string;
  role: string;
  created_at: string;
};

export type PlayerProfile = {
  id: string;
  user_id: string;
  nickname: string;
  region: string;
  languages: string[];
  voice_chat: boolean;
  bio: string;
  created_at: string;
  updated_at: string;
};

export type Game = {
  id: string;
  name: string;
  modes: string[];
  roles: string[];
};

export type Listing = {
  id: string;
  owner_id: string;
  game_id: string;
  title: string;
  mode: string;
  required_roles: string[];
  rank_min: string;
  rank_max: string;
  region: string;
  description: string;
  status: "open" | "closed";
  created_at: string;
  updated_at: string;
};

export type ListingDetails = Listing & {
  game: Game;
  owner: User;
  owner_profile: PlayerProfile;
  applications_count: number;
};

export type Application = {
  id: string;
  listing_id: string;
  applicant_id: string;
  message: string;
  status: "pending" | "accepted" | "rejected";
  created_at: string;
  updated_at: string;
};

export type ApplicationDetails = Application & {
  listing: Listing;
  game: Game;
  applicant: User;
  applicant_profile: PlayerProfile;
};

export type AuthResponse = {
  token: string;
  user: User;
  profile: PlayerProfile;
};
