class Pivot < Formula
  desc "GitHub Issues Management CLI with offline sync capabilities"
  homepage "https://github.com/rhino11/pivot"
  version "1.0.2"
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/rhino11/pivot/releases/download/v1.0.1/pivot-darwin-arm64"
      sha256 "c37ccb349a1cf07490b0481a20d58fd554eb62cf6409e2065a3afe8e1bc65174"
    else
      url "https://github.com/rhino11/pivot/releases/download/v1.0.1/pivot-darwin-amd64"
      sha256 "bc901bc05bd4be1a1e5a17125b787b94d725fb3e12802b187c16605543c4e5c1"
    end
  end

  def install
    bin.install Dir["pivot-darwin-*"].first => "pivot"
  end

  test do
    assert_match "pivot", shell_output("#{bin}/pivot version")
    assert_match "help", shell_output("#{bin}/pivot help")
    
    # Test that binary exists and is executable
    assert_predicate bin/"pivot", :exist?
    assert_predicate bin/"pivot", :executable?
  end
end
