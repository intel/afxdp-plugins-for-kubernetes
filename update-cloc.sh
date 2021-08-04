echo "***   Update CLOC   ***"
if hash cloc 2>/dev/null; then
	cloc $(git ls-files)
	sed -i "/<\!---clocstart--->/,/<\!---clocend--->/c\<\!---clocstart--->\n\`\`\`\n$(cloc $(git ls-files) | sed -n '/-----/,$p' | sed -z 's/\n/\\n/g')\n\`\`\`\n\<\!---clocend--->" README.md
else
	echo "CLOC not installed, skipping"
fi
