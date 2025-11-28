{ lib, buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname = "g-tui-go";
  version = "0.1.0"; #<-- Change this to match your git tags

  src = fetchFromGitHub {
    owner = "garthtrickett";
    repo = pname;
    rev = "31958e42211d23ea47318721f2f815033c4fba43"; #<-- This is the latest commit, change to a tag when you have one
    # To get the sha256 for a new revision, run:
    # nix-prefetch-url --unpack https://github.com/garthtrickett/g-tui-go/archive/<rev>.tar.gz
    sha256 = "19gscz7f154v42ikv35wgqshqg98cnqn2nw0w7q855v5v2b6iij9";
  };

  # To get the vendorSha256 for new dependencies, run:
  # nix-build -A g-tui-go.vendorSha256
  # Or temporarily replace the hash with lib.fakeSha256 and build.
  vendorSha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

  meta = with lib; {
    description = "A terminal UI for interacting with Google's Gemini models";
    homepage = "https://github.com/garthtrickett/g-tui-go";
    license = licenses.mit; #<-- Make sure this matches your project's license
    maintainers = with maintainers; [ "garthtrickett" ];
  };
}
