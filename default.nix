{ lib, buildGoModule, fetchFromGitHub, fetchgit }:

buildGoModule rec {
  pname = "g-tui-go";
  version = "0.1.0"; #<-- Change this to match your git tags

  src = fetchgit {
    url = "https://github.com/garthtrickett/g-tui-go";
    rev = "77801ae14f538693e4e7b6091bfd09cea05f7ab6";
    sha256 = "sha256-zQhqUgFaYVX+Lcq9SDfQazyxEYUZr/zNuGkj6viOfos=";
  };

  vendorHash = "sha256-8A24/OxIub1UUay91PVh+gUsOZJj0pKfR0DOOdrDNPE=";



  meta = with lib; {
    description = "A terminal UI for interacting with Google's Gemini models";
    homepage = "https://github.com/garthtrickett/g-tui-go";
    license = licenses.mit; #<-- Make sure this matches your project's license
    maintainers = with maintainers; [ "garthtrickett" ];
  };
}
